#!/usr/bin/env python3
"""Package the real cmd/arcmd, with compiler resources but without Node.

Run under /workspace/arcmd-maintenance/repo.lock (flock -n), never concurrently
with upstream synchronization. Publishes nothing; release creation is separate.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import plistlib
import shutil
import stat
import tempfile
import zipfile

import package_macos as common

ROOT = common.ROOT
APP = "ArcMD.app"


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--out', type=Path, required=True)
    p.add_argument('--upstream-sha', required=True)
    p.add_argument('--upstream-release-json', type=Path, required=True)
    args = p.parse_args()
    common.run('git', 'diff', '--exit-code')
    common.run('git', 'diff', '--cached', '--exit-code')
    source = common.run('git', 'rev-parse', 'HEAD', capture_output=True, text=True).stdout.strip()
    upstream = json.loads(args.upstream_release_json.read_text())
    prefix = 'ArcMD-macOS-arm64-' + source[:7]
    out = args.out.resolve()
    out.mkdir(parents=True, exist_ok=True)
    archive = out / (prefix + '.zip')
    if archive.exists():
        raise FileExistsError(archive)
    with tempfile.TemporaryDirectory(prefix='arcmd-main-build-') as tmp:
        release = Path(tmp) / prefix
        app = release / APP
        contents = app / 'Contents'
        resources = contents / 'Resources'
        binary = contents / 'MacOS' / 'ArcMD'
        binary.parent.mkdir(parents=True)
        compiler = resources / 'jsx-webview'
        compiler.mkdir(parents=True)
        for name in ('compile.mjs', 'package.json', 'package-lock.json'):
            shutil.copy2(ROOT / 'experiments/jsx-webview' / name, compiler / name)
        # Install *target* native esbuild, not the Linux host executable. No npm
        # lifecycle scripts, downloads at runtime, server or Node runtime bundled.
        common.run('npm', 'ci', '--prefix', str(compiler), '--omit=dev',
                   '--os=darwin', '--cpu=arm64', '--ignore-scripts', '--no-audit', '--no-fund')
        shutil.rmtree(compiler / 'node_modules/.bin', ignore_errors=True)
        native = compiler / 'node_modules/@esbuild/darwin-arm64/bin/esbuild'
        common.macho_details(native.read_bytes())
        # Smoke-test the exact packaged JS modules on Linux. Only this build-time
        # test overrides the esbuild binary; target execution needs a real Mac.
        host_esbuild = ROOT / 'experiments/jsx-webview/node_modules/@esbuild/linux-x64/bin/esbuild'
        manifest = Path(tmp) / 'smoke.json'
        common.run('node', str(compiler / 'compile.mjs'),
                   'experiments/jsx-webview/demo.mdx', str(manifest),
                   env=dict(os.environ, ESBUILD_BINARY_PATH=str(host_esbuild)))
        smoke = json.loads(manifest.read_text())
        if not smoke.get('islands'):
            raise ValueError('packaged compiler produced no islands')
        tool = shutil.which('rcodesign') or str(Path.home() / '.cargo/bin/rcodesign')
        common.run(tool, 'sign', str(native))
        common.macho_details(native.read_bytes())
        env = dict(os.environ, GOOS='darwin', GOARCH='arm64', CGO_ENABLED='0')
        common.run('go', 'build', '-trimpath', '-ldflags', '-s -w',
                   '-o', str(binary), './cmd/arcmd', env=env)
        binary.chmod(0o755)
        plist = {
            'CFBundleName': 'ArcMD', 'CFBundleDisplayName': 'ArcMD',
            'CFBundleIdentifier': 'org.ybbapp.arcmd', 'CFBundleExecutable': 'ArcMD',
            'CFBundlePackageType': 'APPL', 'CFBundleInfoDictionaryVersion': '6.0',
            'CFBundleVersion': '1', 'CFBundleShortVersionString': '0.2.0',
            'CFBundleIconFile': 'icon.icns', 'LSMinimumSystemVersion': '13.0',
            'NSPrincipalClass': 'NSApplication', 'NSHighResolutionCapable': True,
            'NSSupportsAutomaticGraphicsSwitching': True,
            'CFBundleSupportedPlatforms': ['MacOSX'], 'LSArchitecturePriority': ['arm64'],
            'LSApplicationCategoryType': 'public.app-category.productivity',
            'CFBundleDevelopmentRegion': 'zh_CN', 'CFBundleLocalizations': ['zh_CN', 'en'],
            'CFBundleDocumentTypes': [{
                'CFBundleTypeName': 'Markdown / ArcMD Document', 'CFBundleTypeRole': 'Editor',
                'LSHandlerRank': 'Alternate',
                'CFBundleTypeExtensions': ['md', 'markdown', 'mdx', 'arcmd'],
            }],
        }
        (contents / 'Info.plist').write_bytes(plistlib.dumps(plist, sort_keys=True))
        (contents / 'PkgInfo').write_bytes(b'APPL????')
        for src, dst in [('resources/icon.icns', 'icon.icns'), ('LICENSE', 'LICENSE'),
                         ('experiments/jsx-webview/demo.mdx', 'demo.mdx')]:
            shutil.copy2(ROOT / src, resources / dst)
        shutil.copytree(ROOT / 'third_party', resources / 'third_party')
        common.dependency_licenses(resources / 'third_party', './cmd/arcmd')
        shutil.copy2(ROOT / 'experiments/jsx-webview/demo.mdx', release / 'demo.mdx')
        notes = (ROOT / 'experiments/jsx-webview/main-release.zh-CN.md').read_text()
        notes += f'\n源码提交：`{source}`\n\n上游 main：`{args.upstream_sha}`；上游 Release：[{upstream["tag_name"]}]({upstream["html_url"]})（ID {upstream["id"]}）。\n'
        (release / 'README.zh-CN.md').write_text(notes)
        (resources / 'README.zh-CN.md').write_text(notes)
        # Finder's environment may not include Homebrew's Node. This explicit
        # launcher changes only this process environment, never global settings.
        launcher = release / 'Open-with-JSX.command'
        launcher.write_text('''#!/bin/bash
set -eu
HERE="$(cd "$(dirname "$0")" && pwd)"
APP="$HERE/ArcMD.app"
if [ ! -d "$APP" ]; then APP="/Applications/ArcMD.app"; fi
export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"
NODE="${ARCMD_JSX_NODE:-$(command -v node || true)}"
if [ -z "$NODE" ] || [ ! -x "$NODE" ]; then
  echo "需要先安装本机 Apple Silicon Node.js 22+；本包没有捆绑 Node。"
  exit 1
fi
export ARCMD_JSX_NODE="$NODE"
exec "$APP/Contents/MacOS/ArcMD" "$@"
''')
        launcher.chmod(0o755)
        info = {
            'sourceCommit': source, 'target': './cmd/arcmd', 'platform': 'darwin/arm64',
            'upstreamCommit': args.upstream_sha, 'upstreamRelease': upstream,
            'goVersion': common.run('go', 'version', capture_output=True, text=True).stdout.strip(),
            'nodeBundled': False, 'compilerBundled': True, 'compilerTarget': 'darwin-arm64',
            'compilerSmokeTest': 'Linux Node + bundled modules + explicit Linux esbuild override',
            'compilerSmokeIslandCount': len(smoke['islands']),
            'bundleAdHocSignature': True, 'developerIDSigned': False,
            'notarized': False, 'macOSRuntimeTested': False,
            'minimumMacOSVersion': '13.0', 'autoUpdaterConfigured': False,
        }
        (resources / 'BUILD-INFO.json').write_text(json.dumps(info, ensure_ascii=False, indent=2)+'\n')
        _, code_resources = common.sign_app(app)
        info['binaryAnalysis'] = common.macho_details(binary.read_bytes())
        info['binarySHA256'] = common.sha(binary)
        info['bundleCodeResourcesSHA256'] = common.sha(code_resources)
        (release / 'BUILD-INFO.json').write_text(json.dumps(info, ensure_ascii=False, indent=2)+'\n')
        (release / 'SHA256SUMS.txt').write_text(''.join(
            common.sha(f)+'  '+f.relative_to(release).as_posix()+'\n'
            for f in sorted(release.rglob('*')) if f.is_file()))
        with zipfile.ZipFile(archive, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as z:
            for f in sorted(release.rglob('*')):
                if f.is_symlink():
                    raise ValueError('unexpected symlink: '+str(f))
                z.write(f, prefix+'/'+f.relative_to(release).as_posix())
        with zipfile.ZipFile(archive) as z:
            if z.testzip() is not None:
                raise ValueError('ZIP CRC failure')
            for exe in ['Contents/MacOS/ArcMD', 'Contents/Resources/jsx-webview/node_modules/@esbuild/darwin-arm64/bin/esbuild']:
                entry = z.getinfo(prefix+'/'+APP+'/'+exe)
                if not (entry.external_attr >> 16) & stat.S_IXUSR:
                    raise ValueError('lost executable permissions')
                common.macho_details(z.read(entry))
            if not z.read(prefix+'/'+APP+'/Contents/_CodeSignature/CodeResources'):
                raise ValueError('missing bundle signature')
        for name in ['README.zh-CN.md', 'BUILD-INFO.json']:
            shutil.copy2(release / name, out / (prefix+'-'+name))
        archive.with_suffix('.zip.sha256').write_text(common.sha(archive)+'  '+archive.name+'\n')
        print(json.dumps({'archive': str(archive), 'sha256': common.sha(archive), 'bytes': archive.stat().st_size, 'sourceCommit': source}, indent=2))


if __name__ == '__main__':
    main()
