package main

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ciao1019/ArcMD/internal/editor"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestReadManifestValidation(t *testing.T) {
	for _, input := range []string{`{}`, `{"version":2}`, `{"version":1,"blocks":[{"language":"wrong"}]}`, `{"version":1} trailing`, strings.Repeat("x", (16<<20)+1)} {
		if _, err := readManifest(strings.NewReader(input)); err == nil {
			t.Fatal("accepted invalid manifest")
		}
	}
	m, err := readManifest(strings.NewReader(`{"version":1,"namespace":"00000000000000000000000000000000","markdown":"# Native","blocks":[{"language":"arcmd-jsx-00000000000000000000000000000000-0","source":"<Counter />"}]}`))
	if err != nil || len(m.Blocks) != 1 {
		t.Fatal(m, err)
	}
}

// This snapshot deliberately shows source mode: Tester cannot render WebKit.
func TestReaderSnapshot(t *testing.T) {
	r := &reader{editor: editor.New("# Native ArcMD + JSX\n\nThe paragraph remains native.\n\n```arcmd-jsx-0\n<Counter initial={5} />\n```\n\nMore native text after the island.\n"), source: true}
	r.editor.ReadOnly = true
	tt := ui.NewTester(r.view, 1060, 800)
	tt.Frame()
	if dir := os.Getenv("ARCMD_SNAPSHOT"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		file, err := os.Create(filepath.Join(dir, "jsx-reader-source.png"))
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if err := png.Encode(file, tt.Image()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCompiledManifestProjection(t *testing.T) {
	path := os.Getenv("ARCMD_JSX_MANIFEST")
	if path == "" {
		t.Skip("set ARCMD_JSX_MANIFEST to a compiler-produced fixture artifact")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	m, err := readManifest(file)
	if err != nil {
		t.Fatal(err)
	}
	e := editor.New(m.Markdown)
	matched, columns := 0, 0
	for _, b := range e.Doc.Blocks {
		if b.Kind == "columns" {
			columns++
		}
		for _, block := range m.Blocks {
			if b.Kind == "code" && b.Language == block.Language && b.Text() == block.Source {
				matched++
			}
		}
	}
	if len(m.Blocks) != 3 || matched != 3 || columns != 1 {
		t.Fatalf("blocks=%d matched=%d native columns=%d", len(m.Blocks), matched, columns)
	}
}

type placementWebView struct{ ui.NativeWebView }

func TestReaderNativePopoverCoversIsland(t *testing.T) {
	r := &reader{editor: editor.New("# Native ArcMD + JSX\n\nThe paragraph remains native.\n\n```arcmd-jsx-0\n<Counter />\n```\n\nAfter island.\n")}
	r.editor.ReadOnly = true
	view := &placementWebView{}
	r.blocks = map[string]editor.EmbeddedBlock{r.editor.Doc.Blocks[2].ID: {View: view, Height: 240}}
	tt := ui.NewTester(r.view, 1060, 800)
	tt.Frame()
	if tt.WebViewAt(440, 250) != view {
		t.Fatal("missing island before native popover")
	}
	if err := tt.Click("原生浮层"); err != nil {
		t.Fatal(err)
	}
	if tt.WebViewAt(440, 250) != nil {
		t.Fatal("popover did not cover island or intercept pointer")
	}
	if err := tt.Click("切换源码"); err != nil {
		t.Fatal(err)
	}
	if !r.source {
		t.Fatal("source toggle did not work")
	}
}
