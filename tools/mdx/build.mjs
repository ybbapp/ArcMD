import {build} from 'esbuild'
import {fileURLToPath} from 'node:url'
import {resolve} from 'node:path'
import {readFile, writeFile} from 'node:fs/promises'
import {createHash} from 'node:crypto'
const root = fileURLToPath(new URL('../..', import.meta.url))
const dep = resolve(root, 'tools/mdx/node_modules')
const json = async path => JSON.parse(await readFile(resolve(root, path), 'utf8'))
const pkg = await json('tools/mdx/package.json')
const dependencies = {...pkg.dependencies, ...pkg.devDependencies}
for (const [name, pin] of Object.entries(dependencies)) {
  const installed = await json(`tools/mdx/node_modules/${name}/package.json`)
  if (!/^\d+\.\d+\.\d+$/.test(pin) || installed.version !== pin) throw new Error(`Exact dependency pin required: ${name}@${pin}`)
}
const matrix = await json('tools/mdx/compatibility.json')
if (dependencies['@mdx-js/mdx'] !== matrix.target.version) throw new Error('Compatibility target/compiler version mismatch')
const common = {absWorkingDir: root, bundle: true, platform: 'browser', target: 'es2020', legalComments: 'inline'}
await build({...common,
  entryPoints: ['tools/mdx/compiler-entry.mjs'], outfile: 'internal/jsx/assets/mdx-compiler.js', format: 'esm',
  nodePaths: [dep], alias: {'decode-named-character-reference': resolve(dep, 'decode-named-character-reference/index.js')}
})
await build({...common,
  stdin: {contents: "import React from 'react';import {createRoot} from 'react-dom/client';import {jsx,jsxs,Fragment} from 'react/jsx-runtime';export {React,createRoot,jsx,jsxs,Fragment}", resolveDir: resolve(root, 'tools/mdx'), loader: 'js'},
  outfile: 'internal/jsx/assets/react-runtime.js', format: 'iife', globalName: 'ArcMDReact', minify: true,
  define: {'process.env.NODE_ENV': '"production"'}
})
const hash = async path => createHash('sha256').update(await readFile(resolve(root, path))).digest('hex')
const inputs = {}, assets = {}
for (const path of ['tools/mdx/compiler-entry.mjs', 'tools/mdx/build.mjs', 'tools/mdx/package.json', 'tools/mdx/package-lock.json', 'tools/mdx/compatibility.json']) inputs[path] = await hash(path)
for (const name of ['mdx-compiler.js', 'react-runtime.js', 'island-wrapper.js', 'style.css']) assets[name] = await hash('internal/jsx/assets/' + name)
assets['measure.js'] = await hash('internal/jsx/measure.js')
await writeFile(resolve(root, 'internal/jsx/assets/build-info.json'), JSON.stringify({schemaVersion: 1, profile: matrix.profile, compiler: matrix.target, dependencies, inputs, assets}, null, 2) + '\n')
