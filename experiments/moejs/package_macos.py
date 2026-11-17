#!/usr/bin/env python3
"""Cross-build real cmd/arcmd with embedded moejs; no runtime Node resources."""
import argparse,hashlib,importlib.util,json,os
from pathlib import Path
import plistlib,shutil,stat,subprocess,tempfile,zipfile
ROOT=Path(__file__).resolve().parents[2]
spec=importlib.util.spec_from_file_location('preview_package',ROOT/'experiments/jsx-webview/package_macos.py')
c=importlib.util.module_from_spec(spec);spec.loader.exec_module(c)
APP='ArcMD MoeJS Test.app'
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--out',type=Path,required=True);args=p.parse_args()
 c.run('git','diff','--exit-code');c.run('git','diff','--cached','--exit-code')
 commit=c.run('git','rev-parse','HEAD',capture_output=True,text=True).stdout.strip()
 prefix='ArcMD-MoeJS-macOS-arm64-'+commit[:7];out=args.out.resolve();out.mkdir(parents=True,exist_ok=True);archive=out/(prefix+'.zip')
 if archive.exists():raise FileExistsError(archive)
 with tempfile.TemporaryDirectory(prefix='arcmd-moejs-') as scratch:
  tmp=Path(scratch);release=tmp/prefix;app=release/APP;contents=app/'Contents';resources=contents/'Resources';resources.mkdir(parents=True);binary=contents/'MacOS/ArcMD';binary.parent.mkdir()
  # Execute this source's actual app with a fresh source and no Node in PATH.
  host=tmp/'arcmd-check';c.run('go','build','-o',str(host),'./cmd/arcmd')
  fresh=tmp/'fresh.mdx';fresh.write_text('# Packaging smoke\n\n<Callout title="Fresh">\n\n<Counter initial={21+21} />\n\n</Callout>\n')
  check=c.run(str(host),'--jsx-check',str(fresh),env=dict(os.environ,PATH='/no-node',ARCMD_JSX_NODE='/missing',ARCMD_JSX_COMPILER='/missing'),capture_output=True,text=True)
  m=json.loads(check.stdout)
  if len(m['blocks'])!=1 or not m['blocks'][0]['html'] or m['blocks'][0]['error']:raise ValueError('no-Node packaging smoke failed')
  c.run('go','build','-trimpath','-ldflags','-s -w','-o',str(binary),'./cmd/arcmd',env=dict(os.environ,CGO_ENABLED='0',GOOS='darwin',GOARCH='arm64'));binary.chmod(0o755)
  info={'sourceCommit':commit,'target':'./cmd/arcmd','platform':'darwin/arm64','backend':'moejs v0.1.1','nodeBundled':False,'nodeRuntimeRequired':False,'compilerEmbedded':True,'sourcePrecompiled':False,'hostExecutableNoNodeSmoke':True,'macOSRuntimeTested':False,'notarized':False,'developerIDSigned':False,'bundleAdHocSignature':True,'compilerSHA256':c.sha(ROOT/'internal/jsx/assets/mdx-compiler.js'),'goVersion':c.run('go','version',capture_output=True,text=True).stdout.strip()}
  plist={'CFBundleName':'ArcMD MoeJS Test','CFBundleDisplayName':'ArcMD MoeJS Test','CFBundleIdentifier':'org.ybbapp.arcmd.moejs-test','CFBundleExecutable':'ArcMD','CFBundlePackageType':'APPL','CFBundleInfoDictionaryVersion':'6.0','CFBundleVersion':'2','CFBundleShortVersionString':'0.2.0','CFBundleGetInfoString':'moejs-test '+commit[:7],'CFBundleIconFile':'icon.icns','LSMinimumSystemVersion':'13.0','NSPrincipalClass':'NSApplication','NSHighResolutionCapable':True,'CFBundleSupportedPlatforms':['MacOSX'],'LSArchitecturePriority':['arm64'],'CFBundleDocumentTypes':[{'CFBundleTypeName':'Markdown / ArcMD Document','CFBundleTypeRole':'Editor','LSHandlerRank':'Alternate','CFBundleTypeExtensions':['md','markdown','mdx','arcmd']}]}
  (contents/'Info.plist').write_bytes(plistlib.dumps(plist));(contents/'PkgInfo').write_bytes(b'APPL????')
  shutil.copy2(ROOT/'resources/icon.icns',resources/'icon.icns');shutil.copy2(ROOT/'LICENSE',resources/'LICENSE');shutil.copytree(ROOT/'third_party',resources/'third_party');c.dependency_licenses(resources/'third_party','./cmd/arcmd')
  # moejs specifically requires this notice in addition to Apache LICENSE.
  mod=Path(c.run('go','list','-m','-f','{{.Dir}}','github.com/MoeForge/moejs',capture_output=True,text=True).stdout.strip())
  shutil.copy2(mod/'THIRD_PARTY_NOTICES',resources/'third_party/moejs-THIRD_PARTY_NOTICES')
  shutil.copy2(ROOT/'experiments/moejs/demo.mdx',release/'demo.mdx')
  (release/'unknown-component.mdx').write_text('# Intentional render failure\n\n<Missing />\n')
  (release/'invalid.mdx').write_text('# Intentional compile failure\n\n<Counter initial={ />\n')
  notes=(ROOT/'experiments/moejs/README.zh-CN.md').read_text()+'\n源码提交：`'+commit+'`\n'
  (release/'README.zh-CN.md').write_text(notes);(resources/'README.zh-CN.md').write_text(notes)
  (resources/'BUILD-INFO.json').write_text(json.dumps(info,ensure_ascii=False,indent=2)+'\n')
  _,sig=c.sign_app(app);info['binarySHA256']=c.sha(binary);info['binaryAnalysis']=c.macho_details(binary.read_bytes());info['bundleCodeResourcesSHA256']=c.sha(sig)
  (release/'BUILD-INFO.json').write_text(json.dumps(info,ensure_ascii=False,indent=2)+'\n')
  (release/'SHA256SUMS.txt').write_text(''.join(c.sha(f)+'  '+f.relative_to(release).as_posix()+'\n' for f in sorted(release.rglob('*')) if f.is_file()))
  with zipfile.ZipFile(archive,'x',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as z:
   for f in sorted(release.rglob('*')):
    if f.is_symlink():raise ValueError('unexpected symlink')
    z.write(f,prefix+'/'+f.relative_to(release).as_posix())
  with zipfile.ZipFile(archive) as z:
   assert z.testzip() is None
   exe=z.getinfo(prefix+'/'+APP+'/Contents/MacOS/ArcMD');assert (exe.external_attr>>16)&stat.S_IXUSR;c.macho_details(z.read(exe))
   assert z.read(prefix+'/'+APP+'/Contents/_CodeSignature/CodeResources')
   assert not any('/node_modules/' in i.filename or i.filename.endswith('/node') for i in z.infolist())
  for name in ['README.zh-CN.md','BUILD-INFO.json']:shutil.copy2(release/name,out/(prefix+'-'+name))
  archive.with_suffix('.zip.sha256').write_text(c.sha(archive)+'  '+archive.name+'\n')
  print(json.dumps({'archive':str(archive),'sha256':c.sha(archive),'bytes':archive.stat().st_size,'sourceCommit':commit},indent=2))
if __name__=='__main__':main()
