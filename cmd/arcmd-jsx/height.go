package main

import (
	_ "embed"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed island_measure.js
var islandMeasureScript string

const (
	islandHeightMin     = 80
	islandHeightMax     = 2000
	islandHeightSlop    = 2
	islandHeightPollFor = 500 * time.Millisecond
)

var islandHeightSettle = []time.Duration{
	0,
	16 * time.Millisecond,
	50 * time.Millisecond,
	120 * time.Millisecond,
	250 * time.Millisecond,
	500 * time.Millisecond,
}

func parseIslandMeasure(v any) (height float32, ready bool, ok bool) {
	switch n := v.(type) {
	case nil:
		return 0, false, false
	case float64:
		return finiteHeight(n)
	case float32:
		return finiteHeight(float64(n))
	case int:
		return finiteHeight(float64(n))
	case int32:
		return finiteHeight(float64(n))
	case int64:
		return finiteHeight(float64(n))
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false, false
		}
		return finiteHeight(f)
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return 0, false, false
		}
		return finiteHeight(f)
	case map[string]any:
		h, _, parsed := parseIslandMeasure(n["height"])
		if !parsed {
			return 0, false, false
		}
		ready, _ = n["ready"].(bool)
		return h, ready, true
	default:
		return 0, false, false
	}
}

func finiteHeight(n float64) (float32, bool, bool) {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, false, false
	}
	return float32(math.Round(n)), true, true
}

func clampIslandHeight(h float32) float32 {
	if math.IsNaN(float64(h)) || math.IsInf(float64(h), 0) {
		return islandHeightMin
	}
	return min(islandHeightMax, max(islandHeightMin, h))
}

func (r *reader) setIslandHeight(view ui.NativeWebView, height float32) bool {
	if view == nil {
		return false
	}
	height = clampIslandHeight(height)
	for id, block := range r.blocks {
		if block.View != view {
			continue
		}
		if absHeight(block.Height-height) <= islandHeightSlop {
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
