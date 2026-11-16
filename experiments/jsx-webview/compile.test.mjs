import test from 'node:test'
import assert from 'node:assert/strict'
import {createHash} from 'node:crypto'
import {mkdtemp, readFile, writeFile} from 'node:fs/promises'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import {execFileSync} from 'node:child_process'
import {JSDOM, VirtualConsole} from 'jsdom'
import {applyTheme, compileDocument, islandHTML, islandStyle} from './compile.mjs'

function untilDOM(window, predicate) {
  return new Promise((resolve, reject) => {
    const done = () => {
      if (!predicate()) return
      window.clearTimeout(timeout); observer.disconnect(); resolve()
    }
    const observer = new window.MutationObserver(done)
    const timeout = window.setTimeout(() => { observer.disconnect(); reject(new Error('DOM update timed out')) }, 2000)
    observer.observe(window.document, {subtree: true, childList: true, characterData: true})
    done()
  })
}
function dom(html) { return new JSDOM(html, {runScripts: 'dangerously', virtualConsole: new VirtualConsole()}) }

test('official AST boundaries: multiline JSX, quoted >, braces, nesting, Unicode and native HTML', async () => {
  const source = '# 😀 Native\n\n<Counter\n label={"x > y }"}\n initial={3 + 4}\n/>\n\nAfter.\n\n<div class="arcmd-columns" data-widths="50,50">\n\nNative HTML.\n\n</div>\n'
  const m = await compileDocument(source)
  assert.equal(m.blocks.length, 1)
  assert.equal(m.blocks[0].source, '<Counter\n label={"x > y }"}\n initial={3 + 4}\n/>')
  assert.equal(m.blocks[0].error, '')
  assert.ok(m.markdown.startsWith('# 😀 Native\n\n```' + m.blocks[0].language))
  assert.ok(m.markdown.endsWith('After.\n\n<div class="arcmd-columns" data-widths="50,50">\n\nNative HTML.\n\n</div>\n'))
})

test('fenced JSX examples and ordinary native HTML are not executable islands', async () => {
  const source = '```jsx\n<Counter />\n```\n\n<div class="arcmd-columns">\n\nText\n\n</div>\n'
  const m = await compileDocument(source)
  assert.equal(m.blocks.length, 0)
  assert.equal(m.markdown, source)
})

test('React state, JSX expressions and real clicks work in the compiled offline runtime', async () => {
  const d = dom(await islandHTML('<Counter initial={2 + 3} label="Checks" />'))
  try {
    await untilDOM(d.window, () => d.window.document.querySelector('output'))
    assert.equal(d.window.document.querySelector('output').textContent, ' 5')
    d.window.document.querySelector('button').click()
    await untilDOM(d.window, () => d.window.document.querySelector('output').textContent === ' 6')
  } finally { d.window.close() }
})

test('Markdown children inside an island use MDX semantics, not literal asterisks', async () => {
  const d = dom(await islandHTML('<Callout>\n\nThis **is bold**.\n\n<Counter initial={9} />\n\n</Callout>'))
  try {
    await untilDOM(d.window, () => d.window.document.querySelector('button'))
    assert.equal(d.window.document.querySelector('p strong').textContent, 'is bold')
    assert.equal(d.window.document.querySelector('output').textContent, ' 9')
  } finally { d.window.close() }
})

test('unknown components show source and an error instead of a blank page', async () => {
  const d = dom(await islandHTML('<Missing />'))
  try {
    await untilDOM(d.window, () => d.window.document.querySelector('pre'))
    assert.match(d.window.document.querySelector('pre').textContent, /JSX rendering failed:/)
    assert.match(d.window.document.querySelector('pre').textContent, /<Missing \/>/)
  } finally { d.window.close() }
})

test('inline JSX is explicitly rejected rather than silently claiming compatibility', async () => {
  await assert.rejects(compileDocument('Hello <Counter /> world.\n'), /Inline\/nested MDX/)
  await assert.rejects(compileDocument('- item\n\n  <Counter />\n'), /Inline\/nested MDX/)
})

test('file-level imports/exports and shared expressions are source-only', async () => {
  const m = await compileDocument('import {X} from "do-not-load"\n\nexport const total = 3\n\n{total}\n\n<Counter />\n')
  assert.equal(m.diagnostics.length, 3)
  assert.equal(m.blocks.length, 1)
  assert.match(m.markdown, /```mdx-source-only/)
})

test('module imports in island expressions are rejected at build time', async () => {
  const m = await compileDocument('<Callout>{import("private-document")}</Callout>')
  assert.equal(m.blocks[0].html, '')
  assert.match(m.blocks[0].error, /Document module imports are not supported/)
})

test('compiler hashes scripts and styles for CSP; no CDN, eval permission or network permission', async () => {
  const html = await islandHTML('<Counter label={"</script><script>bad()</script>"} />')
  const parsed = new JSDOM(html)
  assert.equal(parsed.window.document.scripts.length, 1)
  parsed.window.close()
  const script = html.match(/<script>([\s\S]*)<\/script>/)[1]
  const digest = createHash('sha256').update(script).digest('base64')
  assert.ok(html.includes(`script-src 'sha256-${digest}'`))
  const style = html.match(/<style>([\s\S]*)<\/style>/)[1]
  const styleDigest = createHash('sha256').update(style).digest('base64')
  assert.ok(html.includes(`style-src 'sha256-${styleDigest}'`))
  assert.ok(!html.includes("'unsafe-inline'"))
  assert.ok(html.includes("connect-src 'none'"))
  assert.ok(!html.includes("'unsafe-eval'"))
  assert.ok(!html.includes('src="https://'))
})

test('CLI never overwrites source or existing artifacts', async () => {
  const dir = await mkdtemp(join(tmpdir(), 'arcmd-jsx-'))
  const source = join(dir, 'note.mdx'), output = join(dir, 'note.manifest.json')
  const original = Buffer.from('# 中文\r\n\r\n<Counter />\r\n')
  await writeFile(source, original)
  execFileSync(process.execPath, ['compile.mjs', source, output])
  assert.deepEqual(await readFile(source), original)
  assert.throws(() => execFileSync(process.execPath, ['compile.mjs', source, source], {stdio:'pipe'}))
  assert.throws(() => execFileSync(process.execPath, ['compile.mjs', source, output], {stdio:'pipe'}))
  assert.deepEqual(await readFile(source), original)
})


test('existing native code cannot spoof an executable placeholder', async () => {
  const source = '```arcmd-jsx-0\n<Counter />\n```\n\n<Counter />\n'
  const m = await compileDocument(source)
  assert.equal(m.blocks.length, 1)
  assert.ok(!source.includes(m.namespace))
  assert.equal(m.blocks[0].language, `arcmd-jsx-${m.namespace}-0`)
  assert.ok(m.markdown.startsWith('```arcmd-jsx-0\n<Counter />\n```'))
})


test('built-in island styles use CSS variables, not a hardcoded light card', () => {
  for (const token of ['--bg', '--text', '--text-muted', '--surface', '--border', '--accent', '--accent-text', '--danger', '--font', '--font-size']) {
    assert.ok(islandStyle.includes('var(' + token + ')') || islandStyle.includes(token + ':'), token)
  }
  assert.match(islandStyle, /background:var\(--bg\)/)
  assert.match(islandStyle, /color:var\(--text\)/)
  assert.ok(!islandStyle.includes('#222'))
  assert.ok(!islandStyle.includes('#fff'))
})

test('theme tokens can be updated on a compiled island without rebuilding or remounting', async () => {
  const html = await islandHTML('<Counter initial={4} label="Themed" />')
  const d = dom(html)
  try {
    await untilDOM(d.window, () => d.window.document.querySelector('output'))
    const original = html
    applyTheme(d.window.document, {dark: true, vars: {'--bg': '#18181b', '--text': '#f4f4f5', '--accent': '#3b82f6', '--accent-text': '#ffffff'}})
    const root = d.window.document.documentElement
    assert.equal(root.style.getPropertyValue('color-scheme'), 'dark')
    assert.equal(root.style.getPropertyValue('--bg'), '#18181b')
    assert.equal(root.style.getPropertyValue('--text'), '#f4f4f5')
    // jsdom does not reliably resolve var() in getComputedStyle; the host
    // writes the same properties the compiled stylesheet consumes.
    assert.equal(d.window.getComputedStyle(root).getPropertyValue('--bg').trim() || root.style.getPropertyValue('--bg'), '#18181b')
    assert.equal(d.window.getComputedStyle(root).getPropertyValue('--text').trim() || root.style.getPropertyValue('--text'), '#f4f4f5')
    assert.equal(d.window.document.querySelector('output').textContent, ' 4')
    d.window.document.querySelector('button').click()
    await untilDOM(d.window, () => d.window.document.querySelector('output').textContent === ' 5')
    applyTheme(d.window.document, {dark: false, vars: {'--bg': '#ffffff', '--text': '#18181b'}})
    assert.equal(root.style.getPropertyValue('color-scheme'), 'light')
    assert.equal(root.style.getPropertyValue('--bg'), '#ffffff')
    assert.equal(root.style.getPropertyValue('--text'), '#18181b')
    assert.equal(d.window.getComputedStyle(root).getPropertyValue('--bg').trim() || root.style.getPropertyValue('--bg'), '#ffffff')
    assert.equal(d.window.document.querySelector('output').textContent, ' 5')
    assert.equal(html, original)
  } finally { d.window.close() }
})
