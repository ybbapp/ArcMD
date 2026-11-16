package app

import (
	"log"
	"time"

	model "github.com/Ciao1019/ArcMD/internal/document"
	"github.com/Ciao1019/ArcMD/internal/editor"
	"github.com/Ciao1019/ArcMD/internal/jsx"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

const jsxHeightPoll = 500 * time.Millisecond

func projectJSX(d *document, source string) *model.Document {
	if d == nil || !jsx.IsMDX(d.rel) {
		return nil
	}
	m, err := jsx.Compile(source, jsx.Options{})
	if err != nil {
		d.jsxErr = err
		return nil
	}
	d.jsx = &m
	d.jsxErr = nil
	return editor.ImportMarkdown(m.Markdown)
}

func (d *document) jsxOverlay() map[string]editor.EmbeddedBlock {
	if d == nil || d.source || len(d.jsxBlocks) == 0 {
		return nil
	}
	return d.jsxBlocks
}

func (d *document) closeJSX() {
	for _, view := range d.jsxViews {
		if w, ok := view.(*mygo.WebView); ok {
			w.Destroy()
		}
	}
	d.jsxViews = nil
	d.jsxBlocks = nil
	d.jsxThemeOK = false
	d.jsxThemeKey = ""
}

func (a *App) jsxWindow(d *document) *mygo.Window {
	if d == nil {
		return a.win
	}
	if w := a.windowOf(d.space.ID, d.rel); w != nil {
		if win, ok := w.win.(*mygo.Window); ok {
			return win
		}
	}
	return a.win
}

func (a *App) newIslandView(d *document) (ui.NativeWebView, error) {
	if a.testIslandView != nil {
		return a.testIslandView(), nil
	}
	win := a.jsxWindow(d)
	if win == nil {
		return nil, nil
	}
	view, err := win.NewWebView(mygo.WebViewOptions{
		Page:              mygo.PageOptions{Untrusted: true},
		ScrollToContainer: true,
		BackgroundColor:   "transparent",
	})
	if err != nil {
		return nil, err
	}
	view.Page().OnWillNavigate(func(event *mygo.NavigateEvent) {
		if event.URL != "about:blank" {
			event.PreventDefault()
		}
	})
	return view, nil
}

func (a *App) ensureJSXViews(d *document) {
	if d == nil || d.jsx == nil || d.source {
		return
	}
	if d.jsxBlocks == nil {
		d.jsxBlocks = make(map[string]editor.EmbeddedBlock)
	}
	for _, block := range d.jsx.Blocks {
		if block.Error != "" || block.HTML == "" {
			continue
		}
		var native *model.Block
		for _, b := range d.ed.Doc.Blocks {
			if b.Kind == "code" && b.Language == block.Language && b.Text() == block.Source {
				native = b
				break
			}
		}
		if native == nil {
			continue
		}
		if existing, ok := d.jsxBlocks[native.ID]; ok && existing.View != nil {
			continue
		}
		view, err := a.newIslandView(d)
		if err != nil {
			log.Printf("jsx island: %v (showing source)", err)
			continue
		}
		if view == nil {
			continue
		}
		height := block.Height
		if height <= 0 {
			height = 240
		}
		if w, ok := view.(*mygo.WebView); ok {
			w.Page().OnDidFinishLoad(func() {
				d.jsxThemeOK = false
				if d.jsxThemeKey != "" {
					_, _ = w.Page().Eval(d.jsxTheme.Script())
					_ = w.SetBackgroundColor(d.jsxTheme.Background())
				}
				a.settleJSXHeight(d, w, 0)
				if win := a.jsxWindow(d); win != nil {
					win.Invalidate()
				}
			})
			w.Page().LoadHTML(block.HTML, "about:blank")
		}
		d.jsxViews = append(d.jsxViews, view)
		d.jsxBlocks[native.ID] = editor.EmbeddedBlock{View: view, Height: height}
	}
}

func (a *App) syncJSXTheme(d *document, t *ui.Theme) {
	if d == nil || t == nil || len(d.jsxViews) == 0 {
		return
	}
	theme := jsx.NewTheme(t)
	key := theme.Key()
	if key != d.jsxThemeKey {
		d.jsxThemeKey = key
		d.jsxTheme = theme
		d.jsxThemeOK = false
		for _, view := range d.jsxViews {
			if w, ok := view.(*mygo.WebView); ok {
				_ = w.SetBackgroundColor(theme.Background())
			}
		}
	}
	if d.jsxThemeOK {
		return
	}
	script := theme.Script()
	ok := true
	for _, view := range d.jsxViews {
		w, is := view.(*mygo.WebView)
		if !is {
			continue
		}
		if _, err := w.Page().Eval(script); err != nil {
			ok = false
		}
	}
	d.jsxThemeOK = ok
}

func (a *App) settleJSXHeight(d *document, view *mygo.WebView, attempt int) {
	delays := []time.Duration{0, 16 * time.Millisecond, 50 * time.Millisecond, 120 * time.Millisecond, 250 * time.Millisecond, 500 * time.Millisecond}
	if d == nil || view == nil || attempt >= len(delays) {
		return
	}
	last := attempt == len(delays)-1
	if a.applyJSXMeasure(d, view, !last) {
		if win := a.jsxWindow(d); win != nil {
			win.Invalidate()
		}
	}
	next := attempt + 1
	if next >= len(delays) {
		return
	}
	time.AfterFunc(delays[next], func() {
		mygo.RunOnMain(func() { a.settleJSXHeight(d, view, next) })
	})
}

func (a *App) applyJSXMeasure(d *document, view *mygo.WebView, requireReady bool) bool {
	if d == nil || view == nil {
		return false
	}
	raw, err := view.Page().Eval(jsx.MeasureScript)
	if err != nil {
		return false
	}
	height, ready, ok := jsx.ParseMeasure(raw)
	if !ok || (requireReady && !ready) {
		return false
	}
	height = jsx.ClampHeight(height)
	for id, block := range d.jsxBlocks {
		if block.View != view {
			continue
		}
		if !jsx.HeightChanged(block.Height, height) {
			return false
		}
		block.Height = height
		d.jsxBlocks[id] = block
		return true
	}
	return false
}

func (a *App) pollJSXHeights(d *document) {
	if d == nil || d.source || len(d.jsxViews) == 0 {
		return
	}
	now := time.Now()
	if !d.jsxLastPoll.IsZero() && now.Sub(d.jsxLastPoll) < jsxHeightPoll {
		return
	}
	d.jsxLastPoll = now
	changed := false
	for _, view := range d.jsxViews {
		w, ok := view.(*mygo.WebView)
		if !ok {
			continue
		}
		if a.applyJSXMeasure(d, w, true) {
			changed = true
		}
	}
	if changed {
		if win := a.jsxWindow(d); win != nil {
			win.Invalidate()
		}
	}
}

func (a *App) syncJSX(c *ui.Context, d *document) {
	if d == nil || d.jsx == nil {
		return
	}
	a.ensureJSXViews(d)
	if c != nil {
		a.syncJSXTheme(d, c.Theme())
	}
	a.pollJSXHeights(d)
}
