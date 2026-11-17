package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ciao1019/ArcMD/internal/jsx"
	"github.com/Ciao1019/ArcMD/internal/vault"
	"github.com/egoist/mygo/ui"
)

type jsxTestView struct{ ui.NativeWebView }

func compilerAvailable(t *testing.T) string {
	t.Helper()
	compiler, err := jsx.FindCompiler()
	if err != nil {
		t.Skip(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(compiler), "node_modules")); err != nil {
		t.Skip("compiler dependencies are not installed")
	}
	return compiler
}

func TestOpenMDXAttachesEmbeddedBlocks(t *testing.T) {
	compilerAvailable(t)
	t.Setenv("ARCMD_JSX_CACHE", t.TempDir())
	a, tt := newTestApp(t)
	a.testIslandView = func() ui.NativeWebView { return &jsxTestView{} }
	source := "# Project\n\nIntro paragraph.\n\n<Counter initial={5} />\n\nAfter the island.\n"
	if err := vault.Write(root(a), "Demo.mdx", source); err != nil {
		t.Fatal(err)
	}
	a.refreshTree()
	a.open("Demo.mdx")
	if a.doc == nil || a.doc.jsxErr != nil || a.doc.jsx == nil {
		t.Fatalf("expected a compiled projection, doc=%v err=%v", a.doc, a.doc.jsxErr)
	}
	if len(a.doc.jsx.Blocks) != 1 || a.doc.jsx.Blocks[0].HTML == "" {
		t.Fatalf("islands: %+v", a.doc.jsx.Blocks)
	}
	tt.Frame()
	if len(a.doc.jsxBlocks) != 1 {
		t.Fatalf("EmbeddedBlock not attached: %+v", a.doc.jsxBlocks)
	}
	var view ui.NativeWebView
	for _, b := range a.doc.jsxBlocks {
		view = b.View
	}
	if view == nil {
		t.Fatal("missing web view")
	}
	// Saving must write the original MDX, not the projection fences.
	if got := a.doc.text(); got != source {
		t.Fatalf("persist text changed the source:\n%s", got)
	}
	if a.doc.dirty() {
		t.Fatal("opening a .mdx file should not mark it dirty")
	}
	disk, err := vault.Read(root(a), "Demo.mdx")
	if err != nil || disk != source {
		t.Fatalf("open rewrote the file: %q %v", disk, err)
	}
}

func TestJSXCompileFailureFallsBackToSource(t *testing.T) {
	t.Setenv("ARCMD_JSX_NODE", "node-not-installed-for-arcmd-jsx")
	t.Setenv("ARCMD_JSX_CACHE", t.TempDir())
	a, tt := newTestApp(t)
	source := "# Broken\n\n<Counter />\n"
	if err := vault.Write(root(a), "Broken.mdx", source); err != nil {
		t.Fatal(err)
	}
	a.refreshTree()
	a.open("Broken.mdx")
	if a.doc == nil || a.doc.loadErr != nil {
		t.Fatal("document should still open")
	}
	if a.doc.jsxErr == nil {
		t.Fatal("expected a compile fallback error")
	}
	if a.doc.jsx != nil || len(a.doc.jsxBlocks) != 0 {
		t.Fatal("failed compile should not attach islands")
	}
	tt.Frame()
	if a.doc.text() != source && !strings.Contains(a.doc.text(), "Broken") {
		t.Fatalf("source fallback lost the file: %q", a.doc.text())
	}
}

func TestJSXThemeAndMeasureScriptsStayExpressions(t *testing.T) {
	if !jsx.ScriptIsExpression(jsx.NewTheme(ui.LightTheme()).Script()) {
		t.Fatal("theme script is not an expression IIFE")
	}
	if !jsx.ScriptIsExpression(jsx.MeasureScript) {
		t.Fatal("measure script is not an expression IIFE")
	}
}

func TestOpenMarkdownWithoutJSXIsUnchanged(t *testing.T) {
	a, _ := newTestApp(t)
	a.open("Welcome.md")
	if a.doc.jsx != nil || a.doc.jsxErr != nil || len(a.doc.jsxBlocks) != 0 {
		t.Fatal("ordinary markdown should not run the JSX compiler")
	}
}
