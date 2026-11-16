package main

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/Ciao1019/ArcMD/internal/editor"
)

func TestParseIslandMeasureAcceptsJSONNumbersAndObjects(t *testing.T) {
	cases := []struct {
		raw   any
		want  float32
		ready bool
		ok    bool
	}{
		{float64(112), 112, true, true},
		{json.Number("112.4"), 112, true, true},
		{json.Number("112.6"), 113, true, true},
		{12, 12, true, true},
		{" 140 ", 140, true, true},
		{map[string]any{"height": float64(96), "ready": true}, 96, true, true},
		{map[string]any{"height": float64(32), "ready": false}, 32, false, true},
		{map[string]any{"ready": true}, 0, false, false},
		{math.NaN(), 0, false, false},
		{math.Inf(1), 0, false, false},
		{nil, 0, false, false},
		{"not-a-number", 0, false, false},
	}
	for _, tc := range cases {
		got, ready, ok := parseIslandMeasure(tc.raw)
		if ok != tc.ok || (ok && (got != tc.want || ready != tc.ready)) {
			t.Fatalf("parseIslandMeasure(%v) = %v, %v, %v; want %v, %v, %v", tc.raw, got, ready, ok, tc.want, tc.ready, tc.ok)
		}
	}
}

func TestClampIslandHeight(t *testing.T) {
	if clampIslandHeight(50) != islandHeightMin || clampIslandHeight(3000) != islandHeightMax {
		t.Fatalf("clamp bounds: %v %v", clampIslandHeight(50), clampIslandHeight(3000))
	}
	if clampIslandHeight(120) != 120 {
		t.Fatal("in-range height should be unchanged")
	}
	if clampIslandHeight(float32(math.NaN())) != islandHeightMin {
		t.Fatal("NaN should fall back to the minimum")
	}
}

func TestIslandMeasureScriptUsesContentBoxNotViewport(t *testing.T) {
	script := islandMeasureScript
	if !strings.Contains(script, "return") {
		t.Fatal("Eval statements must return a value")
	}
	for _, needle := range []string{"getElementById(\"root\")", "getBoundingClientRect", "paddingTop", "paddingBottom", "offsetHeight"} {
		if !strings.Contains(script, needle) {
			t.Fatalf("measure script missing %s", needle)
		}
	}
	if strings.Contains(script, "scrollHeight") && !strings.Contains(script, "Do not use document.documentElement.scrollHeight") {
		t.Fatal("measure script should not read scrollHeight")
	}
	if strings.Contains(script, "height:100%") || strings.Contains(script, "height: 100%") {
		t.Fatal("measure script must not force html/body to the viewport")
	}
}

func TestApplyMeasuredHeightUpdatesEmbeddedBlock(t *testing.T) {
	source := "# Native heading\n\n```arcmd-jsx-0\n<Counter initial={5} />\n```\n\nAfter island.\n"
	r := &reader{editor: editor.New(source), blocks: make(map[string]editor.EmbeddedBlock)}
	r.editor.ReadOnly = true
	view := &placementWebView{}
	id := r.editor.Doc.Blocks[1].ID
	r.blocks[id] = editor.EmbeddedBlock{View: view, Height: 240}
	if !r.applyMeasuredHeight(view, map[string]any{"height": float64(112), "ready": true}) {
		t.Fatal("content height should update the embedded block")
	}
	if r.blocks[id].Height != 112 {
		t.Fatalf("stored height: %v", r.blocks[id].Height)
	}
	if r.applyMeasuredHeight(view, float64(113)) {
		t.Fatal("sub-threshold change should not update height")
	}
	if r.blocks[id].Height != 112 {
		t.Fatal("height changed under the slop threshold")
	}
	if !r.applyMeasuredHeight(view, float64(130)) {
		t.Fatal("later content growth should be applied")
	}
	if r.blocks[id].Height != 130 {
		t.Fatalf("grown height: %v", r.blocks[id].Height)
	}
	if r.applyMeasuredHeight(view, math.NaN()) {
		t.Fatal("invalid measure must not overwrite the current height")
	}
}
