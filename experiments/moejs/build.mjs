import {build} from '../jsx-webview/node_modules/esbuild/lib/main.js'
import {resolve} from 'node:path'
const dep=resolve('experiments/jsx-webview/node_modules')
await build({entryPoints:['experiments/moejs/compiler-entry.mjs'],outfile:'internal/jsx/assets/mdx-compiler.js',bundle:true,format:'esm',platform:'browser',target:'es2020',nodePaths:[dep],alias:{'decode-named-character-reference':resolve(dep,'decode-named-character-reference/index.js')},legalComments:'inline'})
await build({stdin:{contents:"import React from 'react';import {createRoot} from 'react-dom/client';import {jsx,jsxs,Fragment} from 'react/jsx-runtime';export {React,createRoot,jsx,jsxs,Fragment}",resolveDir:resolve('experiments/jsx-webview'),loader:'js'},outfile:'internal/jsx/assets/react-runtime.js',bundle:true,format:'iife',globalName:'ArcMDReact',platform:'browser',target:'es2020',minify:true,define:{'process.env.NODE_ENV':'"production"'},legalComments:'inline'})
