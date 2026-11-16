// Compile only explicitly selected top-level MDX islands. Never evaluate input in Node.
import {compile, createProcessor} from '@mdx-js/mdx'
import {build} from 'esbuild'
import {createHash, randomBytes} from 'node:crypto'
import {readFile, writeFile} from 'node:fs/promises'
import {dirname, resolve} from 'node:path'
import {fileURLToPath, pathToFileURL} from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const hash = text => createHash('sha256').update(text).digest('base64')
const literal = value => JSON.stringify(value).replaceAll('<', '\\u003c').replaceAll('\u2028', '\\u2028').replaceAll('\u2029', '\\u2029')

// Theme tokens are applied at runtime by the native host (Page.Eval is not
// subject to the page CSP). Defaults stay transparent so a dark document is
// not flashed with a white card before the first update.
export const islandStyle = ':root{color-scheme:light;--bg:transparent;--text:CanvasText;--text-muted:GrayText;--surface:Canvas;--border:ButtonBorder;--accent:Highlight;--accent-text:HighlightText;--danger:Maroon;--font:system-ui,sans-serif;--font-size:16px}html,body{margin:0;background:var(--bg);color:var(--text);font:var(--font-size)/1.5 var(--font)}body{padding:16px}button{font:inherit;padding:8px 16px;cursor:pointer;color:var(--accent-text);background:var(--accent);border:1px solid var(--border);border-radius:6px}aside{border-left:4px solid var(--accent);padding:12px;background:var(--surface);color:var(--text)}pre{white-space:pre-wrap;overflow-wrap:anywhere;color:var(--danger)}h3{margin-top:0;color:var(--text)}output{color:var(--text-muted)}'

// Same property writes the Go host performs via Eval. Used by tests to prove
// a compiled island can change theme without being rebuilt or remounted.
export function applyTheme(document, theme) {
  const root = document.documentElement
  if (!root) return false
  root.style.setProperty('color-scheme', theme.dark ? 'dark' : 'light')
  for (const [name, value] of Object.entries(theme.vars || {})) {
    root.style.setProperty(name, value)
  }
  return true
}


export function fence(source, language) {
  const ticks = '`'.repeat(Math.max(3, ...[...source.matchAll(/`+/g)].map(m => m[0].length + 1)))
  return `${ticks}${language}\n${source}\n${ticks}`
}

// Real React state and error boundaries, not JSX-to-HTML string substitution.
function runtime(source) {
  return `
import React from 'react';
import {createRoot} from 'react-dom/client';
import Content from 'arcmd:block';
const source = ${literal(source)};
function fallback(error) {
  const pre = document.createElement('pre');
  pre.textContent = 'JSX rendering failed: ' + String(error?.message || error) + '\\n\\n' + source;
  document.getElementById('root').replaceChildren(pre);
}
window.addEventListener('error', e => fallback(e.error || e.message));
window.addEventListener('unhandledrejection', e => fallback(e.reason));
class Boundary extends React.Component {
  state = {error: null};
  static getDerivedStateFromError(error) { return {error}; }
  render() { return this.state.error ? React.createElement('pre', null, 'JSX rendering failed: ' + this.state.error.message + '\\n\\n' + source) : this.props.children; }
}
function Counter({initial = 0, label = 'Count'}) {
  const [count, setCount] = React.useState(initial);
  return React.createElement('section', null,
    React.createElement('h3', null, label),
    React.createElement('button', {onClick: () => setCount(count + 1)}, 'Increment'),
    React.createElement('output', {'aria-live': 'polite'}, ' ' + count));
}
function Callout({children, title = 'Note'}) {
  return React.createElement('aside', null, React.createElement('strong', null, title), children);
}
createRoot(document.getElementById('root')).render(React.createElement(Boundary, null,
  React.createElement(Content, {components: {Counter, Callout}})));
`
}

export async function islandHTML(source) {
  const compiled = String(await compile(source, {development: false}))
  const result = await build({
    stdin: {contents: runtime(source), resolveDir: here, sourcefile: 'island.js', loader: 'js'},
    bundle: true, write: false, logLevel: 'silent', format: 'iife', platform: 'browser', target: 'es2020',
    minify: true, legalComments: 'inline', define: {'process.env.NODE_ENV': '"production"'},
    plugins: [{name: 'mdx-island', setup(build) {
      build.onResolve({filter: /^arcmd:block$/}, () => ({path: 'block', namespace: 'mdx-island'}))
      build.onLoad({filter: /.*/, namespace: 'mdx-island'}, () => ({contents: compiled, loader: 'js', resolveDir: here}))
      // Do not let document expressions pull packages, local files or URLs into the bundle.
      build.onResolve({filter: /.*/}, args => {
        if (args.importer === 'block' && !['react/jsx-runtime', 'react/jsx-dev-runtime'].includes(args.path)) {
          return {errors: [{text: 'Document module imports are not supported in this experiment'}]}
        }
      })
    }}]
  })
  // Escaping a script end tag preserves JS strings while preventing an HTML escape.
  const script = result.outputFiles[0].text.replace(/<\/script/gi, '<\\/script')
  const style = islandStyle
  const csp = `default-src 'none'; script-src 'sha256-${hash(script)}'; style-src 'sha256-${hash(style)}'; base-uri 'none'; form-action 'none'; frame-src 'none'; connect-src 'none'; object-src 'none'`
  return `<!doctype html><html><head><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="${csp}"><style>${style}</style></head><body><div id="root"></div><script>${script}</script></body></html>`
}

function hasNestedMDX(node) {
  return node.children?.some(child => child.type === 'mdxTextExpression' || child.type === 'mdxFlowExpression' || ((child.type === 'mdxJsxTextElement' || child.type === 'mdxJsxFlowElement') && (child.name === null || /^[A-Z]/.test(child.name))) || hasNestedMDX(child))
}

export async function compileDocument(source) {
  const tree = createProcessor().parse(source)
  // A source document must never be able to turn an existing fenced example
  // into an executable placeholder by guessing the numbered language.
  let namespace
  do { namespace = randomBytes(16).toString('hex') } while (source.includes(namespace))
  const blocks = []
  const diagnostics = []
  let markdown = '', cursor = 0
  for (const node of tree.children) {
    const island = node.type === 'mdxJsxFlowElement' && (node.name === null || /^[A-Z]/.test(node.name))
    const unsupported = node.type === 'mdxjsEsm' || node.type === 'mdxFlowExpression'
    if (!island && !unsupported) {
      // Do not silently pretend that inline JSX or islands in lists/HTML wrappers work.
      if (hasNestedMDX(node)) throw new Error(`Inline/nested MDX is not supported outside an island (line ${node.position.start.line})`)
      continue
    }
    const start = node.position.start.offset, end = node.position.end.offset
    const raw = source.slice(start, end)
    markdown += source.slice(cursor, start)
    if (unsupported) {
      diagnostics.push(`Line ${node.position.start.line}: document imports/exports and shared expressions are source-only`)
      markdown += fence(raw, 'mdx-source-only')
    } else {
      const language = `arcmd-jsx-${namespace}-${blocks.length}`
      const block = {language, source: raw, height: 240, html: '', error: ''}
      try { block.html = await islandHTML(raw) }
      catch (error) { block.error = String(error.message); diagnostics.push(`Line ${node.position.start.line}: ${block.error}`) }
      blocks.push(block)
      markdown += fence(raw, language)
    }
    cursor = end
  }
  markdown += source.slice(cursor)
  return {version: 1, namespace, sourceSha256: createHash('sha256').update(source).digest('hex'), markdown, blocks, diagnostics}
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  const [input, output] = process.argv.slice(2)
  if (!input || !output || resolve(input) === resolve(output)) {
    console.error('Usage: node compile.mjs input.mdx output.manifest.json (different files)')
    process.exitCode = 1
  } else {
    try {
      const manifest = await compileDocument(await readFile(input, 'utf8'))
      await writeFile(output, JSON.stringify(manifest), {flag: 'wx'})
      for (const message of manifest.diagnostics) console.error(message)
      console.log(`Compiled ${manifest.blocks.length} JSX islands; source unchanged`)
    } catch (error) { console.error(String(error.message)); process.exitCode = 1 }
  }
}
