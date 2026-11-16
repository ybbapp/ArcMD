# Native ArcMD + JSX islands: experimental spike

This is **not full JSX/MDX compatibility** and is **not enabled in the ordinary
ArcMD application**. The separate `arcmd-jsx` executable reuses the real
`internal/editor` native canvas, paragraph shaping, semantic HTML columns and
scroll container. Only selected top-level MDX component blocks become system
WebViews. It is not a whole-document HTML renderer or a substitute native UI.

The upstream no-WebView policy is intentionally relaxed only for this explicitly
requested fork experiment. No upstream PR or workflow changes are needed.

## Run

Requirements: Go 1.27.3 (auto toolchain works), Node 22+ (tested with 24), npm.
MyGo v0.7.4 dynamically loads native libraries; no cgo is introduced.
On Linux install WebKitGTK (`libwebkit2gtk-4.1-0` on Debian/Ubuntu,
`webkit2gtk4.1` on Fedora) in addition to MyGo's usual GTK/display dependencies.
Windows needs the WebView2 runtime; macOS uses WKWebView.

From the repository root:

```sh
npm ci --ignore-scripts --prefix experiments/jsx-webview
# Output is exclusive-create: choose a NEW output name on each compilation.
node experiments/jsx-webview/compile.mjs \
  experiments/jsx-webview/demo.mdx /tmp/arcmd-demo-v1.manifest.json
CGO_ENABLED=0 go build -o /tmp/arcmd-jsx ./cmd/arcmd-jsx
/tmp/arcmd-jsx -trust /tmp/arcmd-demo-v1.manifest.json
```

`-trust` is a deliberate consent boundary: review the MDX and compiler-generated
manifest before running it. Do not run downloaded/untrusted documents or manifests.
The demo contains a counter starting at 5, native text, a Callout with real nested
Markdown and its own counter starting at 10, an intentional missing-component
error, the existing **native** semantic-HTML columns, and enough text to scroll.

The app reads only the manifest, does not open or write the source file, and
creates no ArcMD settings/Space/history state. The compiler reads `.mdx` and writes
a NEW disposable JSON artifact. It refuses overwriting the source or an existing
artifact. This artifact is an experimental build product, **not** a replacement
storage format for `.mdx`.

## Implementation

- `compile.mjs`: `@mdx-js/mdx` parses the entire file. Official AST ranges, not a
  tag-matching regex, select top-level capitalized JSX component blocks and
  fragments. A name-capitalization check is only a routing policy, not a parser.
  Ordinary lowercase HTML such as ArcMD columns and fenced examples stay native.
- Each selected block is compiled using the real MDX compiler and bundled with
  React/ReactDOM by esbuild, completely offline at reading time. Built-in `Counter`
  and `Callout` are proof-of-concept components; JSX expressions, native JSX tags
  inside an island and nested Markdown use MDX semantics. Missing components get a
  React error boundary with a visible error and original source.
- The in-memory native projection replaces just the selected ranges with uniquely
  namespaced code placeholders. A random namespace absent from the original
  source prevents fenced examples from spoofing executable placeholders. Go
  additionally matches both language and source.
  It does not save that projection back to `.mdx`.
- `editor.Options.EmbeddedBlocks` maps native block IDs to borrowed WebViews and
  fixed heights. The original source remains shaped for accessibility/copy and
  source-mode fallback. Geometry reserves the rectangle before laying out later
  native paragraphs. WebViews are absolute **siblings** of the writing surface
  in the same native scroll content, not children of the pointer-capturing canvas.
- A view is created once per island and kept for the lifetime of the window.
  Scroll-out and source-mode toggling hide it rather than recreate React state.
  MyGo `ScrollToContainer: true` asks the host to forward wheel events outward.
  Window closure owns WebView destruction. The command logs creation errors and
  leaves the original native source visible when a WebView cannot be created.

## Security boundaries

- `PageOptions.Untrusted: true`: no Go methods (including MyGo window methods),
  no bindings, native shell/file access, permission prompts, downloads or popups.
  Navigation to non-blank URLs is prevented. This harness binds **no Go service**.
- Compiler-produced HTML has hash-based script/style CSP, no network or frame
  permissions, no CDN, no eval permission, no base URL or form actions. Literal
  script end tags are escaped before computing script hashes. Module requests
  originating in document code are rejected during bundling.
- The compiler **does not evaluate document JavaScript in Node**. Expressions
  execute only in the WebView runtime. Nevertheless this is not a security product:
  arbitrary trusted JSX can loop, consume resources, manipulate its own DOM or
  freeze a page. There is no CPU/memory quota. A hand-edited manifest may bypass
  compiler-added CSP; execute only reviewed local build artifacts.
- No PAT, credentials, git URLs containing secrets, remote service or local HTTP
  server is used by the feature. GitHub authentication is outside this code.

## Explicit compatibility limits

| Case | Prototype |
| --- | --- |
| Native Markdown and existing HTML columns | Retained natively |
| Top-level capitalized JSX / fragments | Compiled as independent islands |
| Counter, Callout, JSX expressions, JSX DOM children | Supported inside islands |
| Markdown nested **inside** an island | Compiled by MDX |
| Unknown component / render error | Error + source in web page |
| Bundle/import error | Native source fallback + diagnostic |
| File-level import/export, top-level expressions | Displayed as source only |
| Inline JSX, islands inside lists/HTML wrappers | Explicit compiler error |
| Arbitrary local/npm component imports | Not implemented |
| Shared file scope, context/state across islands | Not implemented |
| Dynamic height / ResizeObserver / host bridge | Not implemented; 240 DIPs |
| File watching / agent edits live refresh | Not implemented; recompile + restart |
| Existing Space/sidebar/file-open integration | Not implemented; separate harness |
| JSX editing, round-trip save, PDF or print of JSX | Not implemented; read-only |
| Vue/Svelte or arbitrary lower-case HTML as an island | Not implemented |

Full MDX compatibility requires resolving file-wide imports/scope, inline layout,
shared context, identity across external edits, resource lifetimes and safe runtime
policy. Multiple independent React roots do not solve those semantics by themselves.

## Tests and GUI acceptance

```sh
npm test --prefix experiments/jsx-webview
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 GOOS=linux go build ./...
CGO_ENABLED=0 GOOS=windows go build ./...
CGO_ENABLED=0 GOOS=darwin go build ./...
ARCMD_SNAPSHOT=/tmp/arcmd-jsx-shots go test ./cmd/arcmd-jsx -run Snapshot -count=1
```

Node tests use jsdom to execute the **real compiled React bundle**, click the
counter, observe state and nested Markdown, and check error fallback. They also
cover source preservation, AST boundaries, import rejection, native HTML routing,
and CSP hash construction. jsdom does **not** validate WebKit or enforce CSP.

Go `ui.Tester` tests exercise real native layout reservation, reflow on a host
height change, source mode, outer native scrolling, view placement/offscreen
hiding, read-only gating, finite height validation and no edit transaction. The
WebView is a placement stub: these tests do **not** execute a platform browser.
The source-mode snapshot also deliberately contains no rendered WebView.

Still required on a GUI machine (do not infer success from compilation):

1. First counter shows 5, increments on click; nested one starts at 10.
2. Native text surrounds the islands without overlap; resize wraps native text
   and keeps web rectangles aligned. Columns still render with native layout.
3. Wheel **over the web page** scrolls the outer native page; clipping at viewport
   edges is correct. Scroll away/back and toggle source twice: state survives.
4. Native popover covers the web content and intercepts clicks; keyboard focus
   moves between native controls and component buttons without a trapped focus.
5. Missing component shows error + source. Disabling WebKit yields native source
   fallback rather than a document rewrite.
6. Open/scroll/toggle/close never changes the SHA-256 of `demo.mdx`.
7. Validate CSP/navigation/bridge denial in the actual system engine before using
   documents beyond the reviewed fixture.

The fork's ordinary `mygo.json` still has Linux WebView packaging disabled. This
harness is built directly with Go and dynamically uses installed WebKitGTK; shipping
an integrated reader will need a separate packaging configuration. Do not claim
that this experimental executable is a packaged production ArcMD release.

## Dependencies and licensing

All npm versions are pinned in `package-lock.json`; no CDN runtime dependencies.
MDX, esbuild, React and ReactDOM are MIT-licensed. jsdom is test-only (MIT).
Their npm distributions retain their license files. Bundled React legal comments
are preserved by esbuild `legalComments: 'inline'`. React and ReactDOM MIT license
copies are also included under `third_party/jsx-experiment/` for redistribution.

## Recorded validation on the implementation box

- Go 1.27.3, Linux amd64; Node 24.22.1; MyGo 0.7.4.
- `npm test`: **11 passed**, including real compiled React state updates in jsdom.
- Native editor and experimental command tests: passed, including compiler/native
  projection of all three demo islands and retention of one native columns block.
- `go vet ./...`, `gofmt -l .`, `git diff --check`: passed.
- `CGO_ENABLED=0` builds of all packages for Linux, Windows and macOS: passed.
- Native source-mode snapshot generated and visually inspected; no claim of a
  screenshot containing rendered system WebViews.
- Unfiltered `go test ./...`: **failed** in six existing `internal/app` tests.
  All six reproduce on the untouched base commit
  `d8f9ee7767b2104246e1207d869765f850efffd9` in a separate baseline directory:
  `TestBottomBlockHandleMenuAndIndentSubmenu`, `TestRebindShortcuts`,
  `TestLibrarySpaces`, `TestLibraryCardDrag`,
  `TestQACitationPreviewDoesNotOccludeAnswerText`, `TestMoreMenuAndSettings`.
  They were not changed or hidden by changing CI/workflows. A second full-suite
  run with only those six exact test names excluded passed across all packages.
- The box has GTK3 but no WebKitGTK/JavaScriptCoreGTK runtime in its dynamic
  linker cache. No GUI/browser was driven or claimed as validated. The native
  host wheel-forwarding, focus, compositing and actual CSP enforcement remain
  GUI acceptance work even though the binaries compile successfully.
