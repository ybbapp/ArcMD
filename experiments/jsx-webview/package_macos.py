#!/usr/bin/env python3
"""Cross-build an isolated, self-contained macOS arm64 demo preview on Linux.

No Apple credentials, external signing service, workflow or GitHub writes.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import plistlib
import shutil
import stat
import struct
import subprocess
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[2]
EXECUTABLE = "ArcMD-JSX-Preview"
APP_NAME = "ArcMD JSX Preview.app"


def run(*args, **kwargs):
    return subprocess.run(args, cwd=ROOT, check=True, **kwargs)


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def macho_details(data):
    magic, cpu, subtype, kind, count, size, flags, reserved = struct.unpack_from("<8I", data)
    if magic != 0xFEEDFACF or cpu != 0x0100000C or kind != 2:
        raise ValueError("not a Mach-O arm64 executable")
    result = {"architecture": "arm64", "format": "Mach-O 64-bit executable", "dylibs": []}
    offset = 32
    for _ in range(count):
        command, length = struct.unpack_from("<II", data, offset)
        if length < 8 or offset + length > 32 + size:
            raise ValueError("invalid load command")
        if command in (0xC, 0x80000018, 0x8000001F, 0x20, 0x80000023):
            name_offset = struct.unpack_from("<I", data, offset + 8)[0]
            name = data[offset + name_offset:offset + length].split(b"\0", 1)[0].decode()
            if not name.startswith(("/usr/lib/", "/System/Library/")):
                raise ValueError("unexpected non-system dylib: " + name)
            result["dylibs"].append(name)
        if command == 0x1D:
            start, signature_size = struct.unpack_from("<II", data, offset + 8)
            sig_magic, sig_length, slots = struct.unpack_from(">III", data, start)
            if sig_magic != 0xFADE0CC0 or sig_length > signature_size:
                raise ValueError("invalid embedded signature")
            for slot in range(slots):
                slot_type, location = struct.unpack_from(">II", data, start + 12 + slot * 8)
                cd = start + location
                if struct.unpack_from(">I", data, cd)[0] == 0xFADE0C02:
                    cd_flags = struct.unpack_from(">I", data, cd + 12)[0]
                    result["codeDirectoryFlags"] = hex(cd_flags)
                    result["adHocCodeSignature"] = bool(cd_flags & 0x2)
        offset += length
    if not result.get("adHocCodeSignature"):
        raise ValueError("missing Go-linker ad-hoc signature for Apple Silicon")
    return result


def dependency_licenses(destination):
    # Go's module list is concatenated JSON objects, not a JSON array.
    text = run("go", "list", "-m", "-json", "all", capture_output=True, text=True).stdout
    decoder = json.JSONDecoder()
    modules = []
    while text.strip():
        module, length = decoder.raw_decode(text.lstrip())
        text = text.lstrip()[length:]
        if module.get("Main"):
            continue
        directory = Path(module["Dir"])
        target = destination / "go-modules" / (module["Path"] + "@" + module["Version"])
        copied = []
        for source in sorted(directory.rglob("*")):
            upper = source.name.upper()
            if source.is_file() and (upper.startswith(("LICENSE", "COPYING", "NOTICE")) or upper == "OFL.TXT"):
                relative = source.relative_to(directory)
                output = target / relative
                output.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(source, output)
                copied.append(str(relative))
        if not copied:
            raise ValueError("no license found for " + module["Path"])
        modules.append({"module": module["Path"], "version": module["Version"], "licenses": copied})
    (destination / "GO-DEPENDENCIES.json").write_text(json.dumps(modules, indent=2) + "\n")


def verify_zip(archive, prefix, digest):
    with zipfile.ZipFile(archive) as z:
        if z.testzip() is not None:
            raise ValueError("ZIP CRC validation failed")
        base = prefix + "/" + APP_NAME + "/Contents/"
        binary = z.getinfo(base + "MacOS/" + EXECUTABLE)
        if not (binary.external_attr >> 16) & stat.S_IXUSR:
            raise ValueError("ZIP lost executable permission")
        macho_details(z.read(binary))
        plist = plistlib.loads(z.read(base + "Info.plist"))
        if plist["CFBundleExecutable"] != EXECUTABLE or plist["CFBundlePackageType"] != "APPL":
            raise ValueError("invalid bundle metadata")
        if hashlib.sha256(z.read(base + "Resources/demo.manifest.json")).hexdigest() != digest:
            raise ValueError("missing or changed bundled demo")
        for name in ("Resources/icon.icns", "Resources/LICENSE", "Resources/demo.mdx", "Resources/BUILD-INFO.json"):
            if not z.read(base + name):
                raise ValueError("empty resource: " + name)
        for item in z.infolist():
            if item.filename.startswith("/") or ".." in Path(item.filename).parts:
                raise ValueError("unsafe ZIP path")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    # Build only committed tracked sources; artifacts live outside the checkout.
    run("git", "diff", "--quiet")
    run("git", "diff", "--cached", "--quiet")
    commit = run("git", "rev-parse", "HEAD", capture_output=True, text=True).stdout.strip()
    prefix = "ArcMD-JSX-Preview-macOS-arm64-" + commit[:7]
    out = args.out.resolve()
    out.mkdir(parents=True, exist_ok=True)
    archive = out / (prefix + ".zip")
    if archive.exists():
        raise FileExistsError("refusing to overwrite " + str(archive))
    with tempfile.TemporaryDirectory(prefix="ArcMD JSX Build ") as scratch:
        release = Path(scratch) / prefix
        contents = release / APP_NAME / "Contents"
        resources = contents / "Resources"
        (contents / "MacOS").mkdir(parents=True)
        resources.mkdir()
        manifest = resources / "demo.manifest.json"
        run("node", "experiments/jsx-webview/compile.mjs", "experiments/jsx-webview/demo.mdx", str(manifest))
        digest = sha(manifest)
        source_hash = sha(ROOT / "experiments/jsx-webview/demo.mdx")
        if json.loads(manifest.read_text())["sourceSha256"] != source_hash:
            raise ValueError("demo source checksum mismatch")
        binary = contents / "MacOS" / EXECUTABLE
        env = dict(os.environ, GOOS="darwin", GOARCH="arm64", CGO_ENABLED="0", GOPROXY="off")
        run("go", "build", "-trimpath", "-ldflags", "-s -w -X=main.bundledDemoSHA256=" + digest,
            "-o", str(binary), "./cmd/arcmd-jsx", env=env)
        binary.chmod(0o755)
        analysis = macho_details(binary.read_bytes())
        plist = {
            "CFBundleName": "ArcMD JSX Preview", "CFBundleDisplayName": "ArcMD JSX Preview",
            "CFBundleIdentifier": "org.ybbapp.arcmd.jsx-preview", "CFBundleExecutable": EXECUTABLE,
            "CFBundlePackageType": "APPL", "CFBundleInfoDictionaryVersion": "6.0",
            "CFBundleVersion": "1", "CFBundleShortVersionString": "0.1.0",
            "CFBundleIconFile": "icon.icns", "LSMinimumSystemVersion": "12.0",
            "NSPrincipalClass": "NSApplication", "NSHighResolutionCapable": True,
            "NSSupportsAutomaticGraphicsSwitching": True,
            "CFBundleSupportedPlatforms": ["MacOSX"], "LSArchitecturePriority": ["arm64"],
            "LSApplicationCategoryType": "public.app-category.developer-tools",
            "CFBundleDevelopmentRegion": "zh_CN", "CFBundleLocalizations": ["zh_CN", "en"],
        }
        (contents / "Info.plist").write_bytes(plistlib.dumps(plist, sort_keys=True))
        (contents / "PkgInfo").write_bytes(b"APPL????")
        shutil.copy2(ROOT / "resources/icon.icns", resources / "icon.icns")
        shutil.copy2(ROOT / "LICENSE", resources / "LICENSE")
        shutil.copytree(ROOT / "third_party", resources / "third_party")
        dependency_licenses(resources / "third_party")
        shutil.copy2(ROOT / "experiments/jsx-webview/demo.mdx", resources / "demo.mdx")
        shutil.copy2(ROOT / "experiments/jsx-webview/demo.mdx", release / "demo.mdx")
        shutil.copy2(ROOT / "experiments/jsx-webview/README.md", resources / "Developer-notes.md")
        notes = (ROOT / "experiments/jsx-webview/preview-release.zh-CN.md").read_text()
        source_url = "https://github.com/ybbapp/ArcMD/commit/" + commit
        notes += "\n源码提交：`" + commit + "`\n\n" + source_url + "\n"
        (release / "README.zh-CN.md").write_text(notes)
        (resources / "README.zh-CN.md").write_text(notes)
        info = {"sourceCommit": commit, "sourceURL": source_url, "platform": "darwin/arm64",
                "goVersion": run("go", "version", capture_output=True, text=True).stdout.strip(),
                "demoManifestSHA256": digest, "demoSourceSHA256": source_hash,
                "binarySHA256": sha(binary), "binaryAnalysis": analysis,
                "developerIDSigned": False, "notarized": False, "macOSRuntimeTested": False,
                "bundleResourceSignature": False, "minimumMacOSVersion": "12.0"}
        (resources / "BUILD-INFO.json").write_text(json.dumps(info, indent=2) + "\n")
        (release / "BUILD-INFO.json").write_text(json.dumps(info, indent=2) + "\n")
        checksums = "".join(sha(p) + "  " + p.relative_to(release).as_posix() + "\n"
                            for p in sorted(release.rglob("*")) if p.is_file())
        (release / "SHA256SUMS.txt").write_text(checksums)
        with zipfile.ZipFile(archive, "x", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as z:
            for path in sorted(release.rglob("*")):
                if path.is_symlink():
                    raise ValueError("unexpected symlink")
                z.write(path, prefix + "/" + path.relative_to(release).as_posix())
        verify_zip(archive, prefix, digest)
        checksum_path = archive.with_suffix(".zip.sha256")
        checksum_path.write_text(sha(archive) + "  " + archive.name + "\n")
        shutil.copy2(release / "README.zh-CN.md", out / (prefix + "-README.zh-CN.md"))
        shutil.copy2(release / "BUILD-INFO.json", out / (prefix + "-BUILD-INFO.json"))
        print(json.dumps({"archive": str(archive), "bytes": archive.stat().st_size,
                          "sha256": sha(archive), "commit": commit, "binaryAnalysis": analysis}, indent=2))


if __name__ == "__main__":
    main()
