// Command arcmd-jsx is an isolated read-only JSX-island experiment using the
// actual ArcMD native editor. It does not change the normal ArcMD application.
package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/Ciao1019/ArcMD/internal/editor"
	"github.com/Ciao1019/ArcMD/internal/i18n"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type island struct {
	Language string  `json:"language"`
	Source   string  `json:"source"`
	HTML     string  `json:"html"`
	Error    string  `json:"error"`
	Height   float32 `json:"height"`
}
type manifest struct {
	Namespace    string   `json:"namespace"`
	Version      int      `json:"version"`
	SourceSHA256 string   `json:"sourceSha256"`
	Markdown     string   `json:"markdown"`
	Blocks       []island `json:"blocks"`
	Diagnostics  []string `json:"diagnostics"`
}

func readManifest(r io.Reader) (manifest, error) {
	var m manifest
	data, err := io.ReadAll(io.LimitReader(r, (16<<20)+1))
	if err != nil {
		return m, err
	}
	if len(data) > 16<<20 {
		return m, errors.New("manifest exceeds 16 MiB")
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, err
	}
	if m.Version != 1 {
		return m, fmt.Errorf("unsupported manifest version %d", m.Version)
	}
	namespace, err := hex.DecodeString(m.Namespace)
	if err != nil || len(namespace) != 16 {
		return m, errors.New("invalid placeholder namespace")
	}
	if len(m.Blocks) > 32 {
		return m, errors.New("prototype supports at most 32 islands")
	}
	for i, block := range m.Blocks {
		if block.Language != fmt.Sprintf("arcmd-jsx-%s-%d", m.Namespace, i) {
			return m, errors.New("invalid island identity")
		}
	}
	return m, nil
}

type reader struct {
	editor       *editor.Editor
	blocks       map[string]editor.EmbeddedBlock
	win          *mygo.Window
	views        []*mygo.WebView
	source, menu bool
	themeKey     string
	themeApplied bool
	theme        islandTheme
	lastPoll     time.Time
}

func (r *reader) view(c *ui.Context) {
	width, _ := c.Size()
	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Padding(8).Gap(12).Children(func() {
			ui.Text(c, i18n.T("ArcMD JSX 实验 — 只读"))
			if ui.Button(c, i18n.T("切换源码")).Clicked() {
				r.source = !r.source
			}
			anchor := ui.Button(c, i18n.T("原生浮层"))
			if anchor.Clicked() {
				r.menu = !r.menu
			}
			ui.Popover(c, anchor, &r.menu, func() {
				ui.Column(c).Width(400).Height(340).Padding(24).Children(func() {
					ui.Text(c, i18n.T("此原生浮层应覆盖网页内容。"))
				})
			})
		})
		blocks := r.blocks
		if r.source {
			blocks = nil
		}
		editor.View(c, r.editor, editor.Options{Width: width, Theme: c.Theme(), EmbeddedBlocks: blocks})
		r.syncIslandTheme(c.Theme())
		r.pollIslandHeights()
	})
}

func (r *reader) syncIslandTheme(t *ui.Theme) {
	if t == nil {
		return
	}
	theme := newIslandTheme(t)
	key := theme.Key()
	if key != r.themeKey {
		r.themeKey = key
		r.theme = theme
		r.themeApplied = false
		for _, view := range r.views {
			_ = view.SetBackgroundColor(theme.Background())
		}
	}
	if r.themeApplied || len(r.views) == 0 {
		return
	}
	script := theme.Script()
	ok := true
	for _, view := range r.views {
		if _, err := view.Page().Eval(script); err != nil {
			ok = false
		}
	}
	r.themeApplied = ok
}

func run() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	m, err := startupManifest(os.Args[1:], executable, bundledDemoSHA256, os.Stdout)
	if errors.Is(err, errShowHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, diagnostic := range m.Diagnostics {
		log.Print(diagnostic)
	}
	r := &reader{editor: editor.New(m.Markdown), blocks: make(map[string]editor.EmbeddedBlock)}
	r.editor.ReadOnly = true
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{Title: "ArcMD JSX islands (experiment)", Width: 1060, Height: 800, Content: ui.View(r.view)})
		r.win = win
		for _, block := range m.Blocks {
			if block.Error != "" || block.HTML == "" {
				continue
			} // Native source fallback.
			for _, native := range r.editor.Doc.Blocks {
				if native.Kind != "code" || native.Language != block.Language || native.Text() != block.Source {
					continue
				}
				// No bound methods, window controls, downloads or permissions.
				// Scroll events belong to the native document, not the page.
				view, err := win.NewWebView(mygo.WebViewOptions{Page: mygo.PageOptions{Untrusted: true}, ScrollToContainer: true, BackgroundColor: "transparent"})
				if err != nil {
					log.Printf("%s: %v (showing source)", block.Language, err)
					continue
				}
				view.Page().OnWillNavigate(func(event *mygo.NavigateEvent) {
					if event.URL != "about:blank" {
						event.PreventDefault()
					}
				})
				// Apply tokens after load without reloading the page later.
				// Measure content height after React paints; do not reset it on theme changes.
				view.Page().OnDidFinishLoad(func() {
					r.themeApplied = false
					if r.themeKey != "" {
						_, _ = view.Page().Eval(r.theme.Script())
						_ = view.SetBackgroundColor(r.theme.Background())
					}
					r.settleIslandHeight(view, 0)
					if r.win != nil {
						r.win.Invalidate()
					}
				})
				view.Page().LoadHTML(block.HTML, "about:blank")
				r.views = append(r.views, view)
				r.blocks[native.ID] = editor.EmbeddedBlock{View: view, Height: block.Height}
			}
		}
		win.Invalidate()
	})
	return mygo.App.Run()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
