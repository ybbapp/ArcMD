package main

import (
	"time"

	"github.com/Ciao1019/ArcMD/internal/jsx"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

const islandHeightPollFor = 500 * time.Millisecond

var islandMeasureScript = jsx.MeasureScript

var islandHeightSettle = []time.Duration{
	0,
	16 * time.Millisecond,
	50 * time.Millisecond,
	120 * time.Millisecond,
	250 * time.Millisecond,
	500 * time.Millisecond,
}

func parseIslandMeasure(v any) (float32, bool, bool) { return jsx.ParseMeasure(v) }

func clampIslandHeight(h float32) float32 { return jsx.ClampHeight(h) }

func (r *reader) setIslandHeight(view ui.NativeWebView, height float32) bool {
	if view == nil {
		return false
	}
	height = clampIslandHeight(height)
	for id, block := range r.blocks {
		if block.View != view {
			continue
		}
		if !jsx.HeightChanged(block.Height, height) {
			return false
		}
		block.Height = height
		r.blocks[id] = block
		return true
	}
	return false
}

func (r *reader) applyMeasuredHeight(view ui.NativeWebView, raw any) bool {
	height, _, ok := parseIslandMeasure(raw)
	if !ok {
		return false
	}
	return r.setIslandHeight(view, height)
}

func (r *reader) measureView(view *mygo.WebView, requireReady bool) bool {
	if view == nil {
		return false
	}
	raw, err := view.Page().Eval(islandMeasureScript)
	if err != nil {
		return false
	}
	height, ready, ok := parseIslandMeasure(raw)
	if !ok || (requireReady && !ready) {
		return false
	}
	return r.setIslandHeight(view, height)
}

func (r *reader) settleIslandHeight(view *mygo.WebView, attempt int) {
	if view == nil || attempt >= len(islandHeightSettle) {
		return
	}
	last := attempt == len(islandHeightSettle)-1
	if r.measureView(view, !last) && r.win != nil {
		r.win.Invalidate()
	}
	next := attempt + 1
	if next >= len(islandHeightSettle) {
		return
	}
	delay := islandHeightSettle[next]
	time.AfterFunc(delay, func() {
		mygo.RunOnMain(func() {
			r.settleIslandHeight(view, next)
		})
	})
}

func (r *reader) pollIslandHeights() {
	if r.source || len(r.views) == 0 {
		return
	}
	now := time.Now()
	if !r.lastPoll.IsZero() && now.Sub(r.lastPoll) < islandHeightPollFor {
		return
	}
	r.lastPoll = now
	changed := false
	for _, view := range r.views {
		if r.measureView(view, true) {
			changed = true
		}
	}
	if changed && r.win != nil {
		r.win.Invalidate()
	}
}

func absHeight(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
