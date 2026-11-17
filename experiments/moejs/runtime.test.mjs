import test from 'node:test'
import assert from 'node:assert/strict'
import {execFileSync} from 'node:child_process'
import {mkdtempSync,writeFileSync,rmSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import {fileURLToPath} from 'node:url'
import {createHash} from 'node:crypto'
import jsdom from '../jsx-webview/node_modules/jsdom/lib/api.js'
const {JSDOM,VirtualConsole}=jsdom
const root=fileURLToPath(new URL('../..',import.meta.url)),scratch=mkdtempSync(join(tmpdir(),'arcmd-moejs-test-')),bin=join(scratch,'arcmd')
execFileSync('go',['build','-o',bin,'./cmd/arcmd'],{cwd:root})
process.on('exit',()=>rmSync(scratch,{recursive:true,force:true}))
function compile(source){const path=join(scratch,'fresh.mdx');writeFileSync(path,source);return JSON.parse(execFileSync(bin,['--jsx-check',path],{env:{...process.env,PATH:'/no-node',ARCMD_JSX_NODE:'/missing',ARCMD_JSX_COMPILER:'/missing'},maxBuffer:32*1024*1024}))}
function dom(html){return new JSDOM(html,{runScripts:'dangerously',virtualConsole:new VirtualConsole()})}
function until(w,f){return new Promise((resolve,reject)=>{const ob=new w.MutationObserver(done);const timer=w.setTimeout(()=>{ob.disconnect();reject(new Error('DOM timeout'))},2000);function done(){if(f()){w.clearTimeout(timer);ob.disconnect();resolve()}}ob.observe(w.document,{subtree:true,childList:true,characterData:true});done()})}
test('actual executable with no Node compiles arbitrary nested MDX and renders interactive React',async()=>{
 const m=compile('# New 😀\n\n<Callout title="New file">\n\nThis **is bold**.\n\n<Counter initial={17+24} />\n\n</Callout>\n')
 assert.equal(m.blocks.length,1);assert.equal(m.blocks[0].error,'')
 const d=dom(m.blocks[0].html);try{await until(d.window,()=>d.window.document.querySelector('output'));assert.equal(d.window.document.querySelector('output').textContent,' 41');assert.equal(d.window.document.querySelector('p strong').textContent,'is bold');d.window.document.querySelector('button').click();await until(d.window,()=>d.window.document.querySelector('output').textContent===' 42');d.window.document.documentElement.style.setProperty('--bg','#111');assert.equal(d.window.document.querySelector('output').textContent,' 42')}finally{d.window.close()}
})
test('unknown component visibly fails with original source',async()=>{const m=compile('<Missing />\n');const d=dom(m.blocks[0].html);try{await until(d.window,()=>d.window.document.querySelector('pre'));assert.match(d.window.document.querySelector('pre').textContent,/JSX rendering failed:/);assert.match(d.window.document.querySelector('pre').textContent,/<Missing/)}finally{d.window.close()}})
test('CSP hashes and mixed-case script-end escaping survive new compiler',()=>{const m=compile('<Counter label={"</SCRIPT><script>bad()</script>"} />\n');const html=m.blocks[0].html;const d=new JSDOM(html);assert.equal(d.window.document.scripts.length,1);d.window.close();const script=html.match(/<script>([\s\S]*)<\/script>/)[1];assert.ok(html.includes("script-src 'sha256-"+createHash('sha256').update(script).digest('base64')+"'"));assert.ok(html.includes("connect-src 'none'"));assert.ok(!html.includes("'unsafe-eval'"))})
