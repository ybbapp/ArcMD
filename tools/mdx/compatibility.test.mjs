import test from 'node:test'
import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import {createHash} from 'node:crypto'
import {compileSync} from '@mdx-js/mdx'
import {JSDOM, VirtualConsole} from 'jsdom'
import {compileDocument, islandHTML, applyTheme} from './test-helpers.mjs'
const matrix = JSON.parse(readFileSync(new URL('./compatibility.json', import.meta.url)))
const pkg = JSON.parse(readFileSync(new URL('./node_modules/@mdx-js/mdx/package.json', import.meta.url)))
const build = JSON.parse(readFileSync(new URL('../../internal/jsx/assets/build-info.json', import.meta.url)))
function dom(html) { return new JSDOM(html, {runScripts: 'dangerously', virtualConsole: new VirtualConsole()}) }
function until(window, predicate) {
  return new Promise((resolve, reject) => {
    const observer = new window.MutationObserver(done)
    const timer = window.setTimeout(() => {observer.disconnect(); reject(new Error('DOM timeout'))}, 2000)
    function done() {if (predicate()) {window.clearTimeout(timer);observer.disconnect();resolve()}}
    observer.observe(window.document, {subtree: true, childList: true, characterData: true, attributes: true})
    done()
  })
}
test('exact compiler target, installed package and generated metadata agree', () => {
  assert.equal(pkg.version, '3.1.1')
  assert.deepEqual(build.compiler, matrix.target)
  assert.equal(build.profile, matrix.profile)
  for (const [path, hash] of Object.entries(build.inputs)) {
    assert.equal(createHash('sha256').update(readFileSync(new URL('../../' + path, import.meta.url))).digest('hex'), hash, path)
  }
})
for (const fixture of matrix.cases) {
  test(`MDX ${matrix.target.version} / ${fixture.id} / host ${fixture.status}`, async () => {
    if (fixture.officialSyntax === false) assert.throws(() => compileSync(fixture.source))
    else assert.doesNotThrow(() => compileSync(fixture.source), 'fixture must be valid in the pinned official compiler')
    if (fixture.compileError) {
      await assert.rejects(compileDocument(fixture.source), new RegExp(fixture.compileError))
      return
    }
    const manifest = await compileDocument(fixture.source)
    assert.equal(manifest.compiler.version, matrix.target.version)
    assert.equal(manifest.compiler.profile, matrix.profile)
    assert.equal(manifest.blocks.length, fixture.blocks)
    if (fixture.diagnostic) assert.match(manifest.diagnostics.join('\n'), new RegExp(fixture.diagnostic))
    if (fixture.blockError) {
      assert.match(manifest.blocks[0].error, new RegExp(fixture.blockError))
      assert.equal(manifest.blocks[0].html, '')
      return
    }
    if (!fixture.blocks || fixture.compileOnly) return
    assert.equal(manifest.blocks[0].error, '')
    const d = dom(manifest.blocks[0].html)
    try {
      if (fixture.runtimeError) {
        await until(d.window, () => d.window.document.querySelector('pre'))
        assert.match(d.window.document.querySelector('pre').textContent, new RegExp(fixture.runtimeError))
        assert.match(d.window.document.querySelector('pre').textContent, /JSX rendering failed/)
      } else {
        await until(d.window, () => d.window.document.querySelector(fixture.selector)?.textContent === fixture.text)
        if (fixture.click) {
          d.window.document.querySelector(fixture.selector).click()
          await until(d.window, () => d.window.document.querySelector(fixture.selector)?.textContent === fixture.textAfter)
        }
      }
    } finally {d.window.close()}
  })
}
test('Tabs and Accordion interact and keep state across theme updates', async () => {
  const html = await islandHTML('<Tabs>\n<TabItem label="First">\n<Counter initial={7} />\n</TabItem>\n<TabItem label="Second">\n<Accordion title="More">\n**Details**\n</Accordion>\n</TabItem>\n</Tabs>\n')
  const d = dom(html)
  try {
    const doc = d.window.document
    await until(d.window, () => doc.querySelector('output'))
    doc.querySelector('section button:not([role])').click()
    await until(d.window, () => doc.querySelector('output').textContent === ' 8')
    applyTheme(doc, {dark: true, vars: {'--bg': '#18181b', '--text': '#ffffff'}})
    assert.equal(doc.querySelector('output').textContent, ' 8')
    doc.querySelectorAll('[role=tab]')[1].click()
    await until(d.window, () => doc.querySelector('summary'))
    doc.querySelector('summary').click()
    await until(d.window, () => doc.querySelector('details').open)
    applyTheme(doc, {dark: false, vars: {'--bg': '#ffffff', '--text': '#18181b'}})
    assert.equal(doc.querySelector('details').open, true)
    assert.equal(doc.querySelectorAll('[role=tab]')[1].getAttribute('aria-selected'), 'true')
    doc.querySelectorAll('[role=tab]')[1].dispatchEvent(new d.window.KeyboardEvent('keydown', {key: 'ArrowLeft', bubbles: true}))
    await until(d.window, () => doc.querySelectorAll('[role=tab]')[0].getAttribute('aria-selected') === 'true')
  } finally {d.window.close()}
})
