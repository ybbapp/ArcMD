
const {React,createRoot} = ArcMDReact;
const Content = (function(){__COMPILED_BODY__})(ArcMDReact).default;
const source = __SOURCE_JSON__;
function fallback(error) {
  document.documentElement.dataset.islandError = String(error?.message || error);
  const pre = document.createElement('pre');
  pre.textContent = 'JSX rendering failed: ' + String(error?.message || error) + '\\n\\n' + source;
  document.getElementById('root').replaceChildren(pre);
}
// Host Page.Eval first runs code as an expression. A SyntaxError there, on an
// about:blank page, is often sanitized to message "Script error." with no
// Error object and an empty filename. That must not unmount a working tree.
// Real component failures still have an Error, or go through the boundary.
function isSanitizedScriptError(event) {
  return event && !event.error && (event.message === 'Script error.' || !event.filename);
}
window.addEventListener('error', e => {
  if (isSanitizedScriptError(e)) return;
  fallback(e.error || e.message);
});
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
// These host components are implemented locally, not copied from another reader.
function TabItem({children}) { return React.createElement(React.Fragment, null, children); }
function Tabs({children, defaultIndex = 0, label = 'Tabs'}) {
  const items = React.Children.toArray(children).filter(child => React.isValidElement(child) && child.type === TabItem);
  const [selected, setSelected] = React.useState(Number.isInteger(defaultIndex) ? defaultIndex : 0);
  const id = React.useId();
  if (!items.length) throw new Error('Tabs requires direct TabItem children');
  const active = Math.max(0, Math.min(items.length - 1, selected));
  function key(event, index) {
    const next = event.key === 'ArrowRight' ? (index + 1) % items.length : event.key === 'ArrowLeft' ? (index + items.length - 1) % items.length : event.key === 'Home' ? 0 : event.key === 'End' ? items.length - 1 : null;
    if (next === null) return;
    event.preventDefault(); setSelected(next);
    event.currentTarget.parentElement.querySelectorAll('[role="tab"]')[next].focus();
  }
  return React.createElement('section', {className: 'mdx-tabs'},
    React.createElement('div', {role: 'tablist', 'aria-label': label}, items.map((item, index) =>
      React.createElement('button', {type: 'button', role: 'tab', key: index, id: id + '-tab-' + index, 'aria-selected': index === active, 'aria-controls': id + '-panel', tabIndex: index === active ? 0 : -1, onClick: () => setSelected(index), onKeyDown: event => key(event, index)}, item.props.label || 'Tab ' + (index + 1)))),
    React.createElement('div', {role: 'tabpanel', id: id + '-panel', 'aria-labelledby': id + '-tab-' + active, tabIndex: 0}, items[active].props.children));
}
function Accordion({children, title = 'Details', defaultOpen = false}) {
  const [open, setOpen] = React.useState(Boolean(defaultOpen));
  return React.createElement('details', {className: 'mdx-accordion', open},
    React.createElement('summary', {'aria-expanded': open, onClick: event => {event.preventDefault(); setOpen(!open);}}, title),
    React.createElement('div', null, children));
}
createRoot(document.getElementById('root')).render(React.createElement(Boundary, null,
  React.createElement(Content, {components: {Counter, Callout, Tabs, TabItem, Accordion}})));
