import {compileSync, createProcessor} from '@mdx-js/mdx'

function fence(source, language) {
  const ticks = '`'.repeat(Math.max(3, ...[...source.matchAll(/`+/g)].map(m => m[0].length + 1)))
  return `${ticks}${language}\n${source}\n${ticks}`
}
function jsxElement(node) {
  return node.type === 'mdxJsxTextElement' || node.type === 'mdxJsxFlowElement'
}
function activeMDX(node) {
  if (node.type === 'mdxTextExpression' || node.type === 'mdxFlowExpression') return true
  if (jsxElement(node)) {
    if (node.name === null || /^[A-Z]/.test(node.name) || node.name.includes('.')) return true
    if (node.attributes?.some(a => a.type === 'mdxJsxExpressionAttribute' || a.value?.type === 'mdxJsxAttributeValueExpression' || /^(className|htmlFor|on[A-Z])/.test(a.name || ''))) return true
  }
  return node.children?.some(activeMDX) || false
}
function contains(node, predicate) {
  if (!node || typeof node !== 'object') return false
  if (predicate(node)) return true
  return Object.values(node).some(value => Array.isArray(value) ? value.some(n => contains(n, predicate)) : contains(value, predicate))
}
function unsupportedRuntime(node) {
  if (contains(node, n => n.type === 'ImportExpression' || n.type === 'ImportDeclaration')) return 'Document module imports are unsupported'
  if (contains(node, n => n.type === 'AwaitExpression')) return 'Await expressions are unsupported by the synchronous island runtime'
  if (contains(node, n => n.type === 'MetaProperty' && n.meta?.name === 'import')) return 'import.meta requires a module loader and is unsupported'
  return ''
}
export function compileMDX(input) {
  const {source, namespace} = JSON.parse(input)
  const tree = createProcessor().parse(source)
  let markdown = '', cursor = 0
  const blocks = [], diagnostics = []
  for (const node of tree.children) {
    // Prose remains native unless this paragraph actually needs MDX evaluation.
    // A paragraph is the smallest native block we can replace without a second
    // inline layout engine; list/quote/heading promotion is deliberately absent.
    const island = node.type === 'mdxFlowExpression' ||
      (node.type === 'mdxJsxFlowElement' && activeMDX(node)) ||
      (node.type === 'paragraph' && activeMDX(node))
    const sourceOnly = node.type === 'mdxjsEsm'
    if (!island && !sourceOnly) {
      if (activeMDX(node)) throw new Error('Nested MDX outside a top-level paragraph or JSX island is unsupported')
      continue
    }
    const start = node.position.start.offset, end = node.position.end.offset
    const raw = source.slice(start, end)
    markdown += source.slice(cursor, start)
    if (sourceOnly) {
      diagnostics.push('Imports/exports/document scope are source-only; no module loader is provided')
      markdown += fence(raw, 'mdx-source-only-' + namespace)
    } else {
      if (blocks.length >= 32) throw new Error('Maximum 32 islands')
      const language = `arcmd-jsx-${namespace}-${blocks.length}`
      const block = {language, source: raw, height: 240, html: '', error: '', compiled: ''}
      try {
        const unsupported = unsupportedRuntime(node)
        if (unsupported) throw new Error(unsupported)
        block.compiled = String(compileSync(raw, {development: false, outputFormat: 'function-body'}))
      } catch (error) {
        block.error = String(error.message)
        diagnostics.push(block.error)
      }
      blocks.push(block)
      markdown += fence(raw, language)
    }
    cursor = end
  }
  markdown += source.slice(cursor)
  return {version: 1, namespace, markdown, blocks, diagnostics}
}
