# MDX 3.1.1: versioned, partial host compatibility

This fork targets **`@mdx-js/mdx` 3.1.1**, with the explicit host profile
`arcmd-mdx3-phase1`. This is an exact compiler dependency and a tested subset,
**not a claim of complete MDX 3 conformance**. Phase 1 first shipped in fork 0.2.3 after
PR #3 was merged and remains in 0.2.4, which also packages official upstream v0.2.1. The
separate upstream contribution remains unchanged. Existing 0.2.2 binaries do not contain
these MDX additions.

Official sources: [pinned 3.1.1 package](https://github.com/mdx-js/mdx/blob/3.1.1/packages/mdx/package.json),
[syntax](https://mdxjs.com/docs/what-is-mdx/), [MDX 3 changes](https://mdxjs.com/blog/v3/).
MDX includes Markdown, JSX, expressions and ESM. Version 3 also accepts await
and ES2024 syntax; accepting a syntax tree is not the same as supplying a
module loader, component registry or matching browser runtime.

## Three distinct compatibility layers

1. **Official grammar:** fixtures are compiled by the exact official 3.1.1
   package in development tests. That reference compiler never replaces the
   actual runtime compiler under test.
2. **ArcMD host:** only supported source blocks/paragraphs are projected into
   islands. Modules, document scope and host resources are deliberately absent.
3. **Browser execution:** actual HTML produced by the Go application is checked
   in JSDOM. Real WKWebView/OS/browser JS features and layout still need native
   validation; a modern syntax accepted by Acorn is not universally supported
   by older macOS WebKit. The compiler bundle's esbuild target is not a
   transpilation target for arbitrary user document expressions.

## Phase 1 capability matrix

| Capability | Host status | What is actually implemented |
| --- | --- | --- |
| Ordinary native Markdown/code fences | Supported, unchanged | Does not invoke a browser for ordinary prose; fenced JSX examples stay inert. |
| Top-level JSX/fragments; Markdown inside an island | Supported | Existing Counter/Callout support retained. |
| Standalone flow expressions | Supported, new | Numbers, strings, conditionals, array/map JSX and IIFEs, with scope local to that island. |
| Lowercase JSX containing expressions/components or JSX-specific props | Supported, new | Expression/spread props, `className`, nesting and event props use MDX/React. Plain native HTML conventions remain native. |
| Inline expressions and JSX in top-level paragraphs | **Partial, new** | The **entire affected paragraph** becomes one WebView; not native per-run inline layout/editing. Ordinary paragraphs remain native. |
| Counter/Callout/Tabs/TabItem/Accordion | Host built-ins | Tabs/TabItem/Accordion are new local implementations, not evidence of MDX language conformance. |
| Empty/comment-only expressions | Partial | Source compiles and round-trips; an empty island still has native minimum height/padding. No zero-height-layout claim. |
| JSX/expressions in headings, lists or quotes outside an island | Unsupported | Explicit boundary error rather than silently claiming compatibility. Wrap intended content in a supported top-level island if appropriate. |
| ESM imports/exports and document-level bindings | Unsupported | Source-only with diagnostics; exported names are not available to independently compiled islands. |
| Dynamic imports, import.meta, await | Unsupported | Explicit compile diagnostics. No file/base-URL/module resolver or async island runtime is provided. |
| Cross-island state, user component files, third-party imports | Unsupported | Known built-ins only; unknown names fail visibly. |
| Full CommonMark/plugin parity inside islands | Partial | ArcMD native serialization/features remain unchanged. Native math/frontmatter/etc. are not automatically MDX compiler plugins; Markdown references defined outside a promoted block are not resolved into it. |
| External images/assets/navigation | Unsupported in islands | Existing hash CSP and navigation restrictions remain in force; promoted paragraph links do not gain native-shell link behavior. |

The machine-readable matrix is `tools/mdx/compatibility.json`. Its cases have
separate official-syntax acceptance, host status, compilation/diagnostic and
browser expectations. Negative/unsupported cases **must reject or diagnose**;
a passing matrix does not mean those features became supported. This is a
curated regression profile, not the complete upstream MDX test suite.

## Design: smallest native block, not a whole-page WebView

`internal/editor.EmbeddedBlock` places views on code blocks; the native model
does not offer embedded inline runs. The compiler therefore recognizes active
MDX in a top-level paragraph and projects that paragraph into one namespaced
code block. Native headings, lists and unaffected paragraphs remain native.
Promoting lists/headings/the whole document, or replacing the editor, is not
part of this phase. Document-wide ESM scope would require a separate module/
execution model; copying exported variables into each island would incorrectly
pretend that state, identity and evaluation order were shared, so we do not do it.

Read `tools/mdx/example.mdx`. Open a local `.mdx` from a space/sidebar/system
file-open action. To change JSX or a promoted paragraph, ask an agent or use an
external editor to edit the file, then reload it. ArcMD's existing Markdown
source view is a **read-only export preview**, not a newly added source editor.
The reload action is available for a clean document; save/resolve native edits
before retrying. Source-mode inspection shows original JSX, not projection fences.

Projection retains exact island source, uses a fresh namespace/source SHA256,
and reverses fences for synchronous and asynchronous native saves. There is
no persistent user-HTML cache. Native Markdown formatting/serialization can
still normalize surrounding Markdown; this is not an arbitrary whole-document
byte-preservation promise. Save/reload is tested for new inline/expression forms.
Theme changes update CSS variables without reloading or resetting component
state; ResizeObserver/measurement IIFEs remain unchanged. Reloads can reset state;
inactive Tabs panels are unmounted and their state is not promised to survive
tab switches. No normal-state implementation banner is added.

## Build identity and reproducibility

Direct npm dependencies remain exact pins (MDX 3.1.1, React/ReactDOM 19.2.0,
esbuild 0.25.12, test-only JSDOM 27.2.0); the lockfile pins transitives. No
"latest" upgrade is made. Normal application execution and Go builds require
**no external Node/npm/esbuild**, and no Node runtime is bundled or local server
started. Node 22.12+ or Node 24 is a **developer** dependency for regeneration/tests:

```sh
npm ci --prefix tools/mdx
npm run build --prefix tools/mdx
npm test --prefix tools/mdx

go test ./internal/jsx ./internal/editor ./cmd/arcmd
go test ./internal/app -run 'Test(JSX|MDX3|OpenMDX|OpenMoe|OpenMarkdownWithoutJSX)'
go vet ./...
```

`build.mjs` writes the two JS assets and deterministic `assets/build-info.json`:
compiler name/version/profile, exact direct dependencies, input/lock SHA256s and
all embedded JS/CSS/measurement hashes. Go validates the actual embedded asset
hashes before compiling a document; mixed/stale assets cause a visible compiler
error, not a version badge pretending success. Tests validate source/lock input
hashes. The windowless diagnostic manifest exposes the actual build identity:

```sh
go run ./cmd/arcmd --jsx-check tools/mdx/example.mdx
# JSON includes compiler.name, compiler.version, compiler.profile,
# compiler.buildSha256, sourceSha256, blocks and diagnostics.
```

The diagnostic command only reads the source. Syntax/partial-source-only errors
return nonzero (partial JSON may still be emitted). Runtime component errors
occur later in the browser and remain visible with their source. The old
`experiments/moejs/build.mjs` forwards to canonical `tools/mdx/build.mjs` to avoid
regenerating a stale compiler. Historical standalone experiment tests remain
historical; they are not substituted for the new actual-Go compatibility suite.

## Security and validation boundary

Only trusted compiler code executes in moejs, with document text as data. No
Go/OS/file/network APIs or module loader are exposed. JS allocation budget
256 MiB, evaluation deadline capped at 5 seconds (initial fixed compiler parsing
precedes it), 512 KiB source, 32 islands, 16 MiB / 200,000-node result caps remain.
These are not process-wide RSS limits or browser-expression CPU/memory limits.
Document expressions run only in Untrusted WebViews, without Go bindings, with
hash-only script/style CSP, network resources denied and navigation restricted
to about:blank. No CSP relaxation or global Linux WebView enablement is introduced.

Tests use the real binary with `PATH=/no-node`, fresh source, browser interaction,
known failure paths, source preservation and explicit unsupported-version features.
JSDOM does not validate real WKWebView layout, CSP enforcement or Finder startup.
The user accepted the phase-1 test package at `2283afd` on their Mac. The new
0.2.4 main package has not itself been tested on a real Mac. The existing six Linux `internal/app`
baseline test failures are not fixed or relabelled as passing.

## Next stage (not included)

A reviewed, bounded document-scope/module model with explicit resolver policy;
resource handling; heading/list/quote integration and native inline layout;
async/await semantics; plugin/reference context; broader official MDX fixtures;
real platform/browser-engine compatibility and security review. These are
necessary before claiming a wider version-conformance level.
