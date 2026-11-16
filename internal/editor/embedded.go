package editor

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// EmbeddedBlock places a host-owned native web view in a read-only code block.
// Height is fixed in DIPs in this prototype; the page cannot call native code.
// Hosts must create views once, not once per frame, and destroy them on close.
type EmbeddedBlock struct {
	View   ui.NativeWebView
	Height float32
}

func (e *Editor) syncEmbeddedBlocks(blocks map[string]EmbeddedBlock) {
	// Keep the ordinary native editor's per-frame path allocation-free.
	if (!e.ReadOnly || len(blocks) == 0) && len(e.embeddedBlocks) == 0 {
		return
	}
	next := make(map[string]EmbeddedBlock)
	if e.ReadOnly {
		for id, block := range blocks {
			leaf, ok := e.Engine.LeafByID(id)
			if !ok || leaf.Block.Kind != "code" || block.View == nil {
				continue
			}
			if math.IsNaN(float64(block.Height)) || math.IsInf(float64(block.Height), 0) {
				block.Height = 240
			}
			block.Height = min(2000, max(80, block.Height))
			next[id] = block
		}
	}
	changed := len(next) != len(e.embeddedBlocks)
	for id, block := range next {
		old, ok := e.embeddedBlocks[id]
		changed = changed || !ok || old.Height != block.Height
	}
	if changed {
		e.ready = false
	}
	e.embeddedBlocks = next
}

func (e *Editor) embeddedBlockViews(c *ui.Context) {
	if e.printing {
		return
	}
	for _, r := range e.rows {
		block, ok := e.embeddedBlocks[r.leaf.Block.ID]
		if !ok {
			continue
		}
		// Siblings of the input surface, inside the same scroll content box:
		// MyGo handles clipping, offscreen hiding and pointer delivery.
		ui.WebView(c.Key("embedded-"+r.leaf.Block.ID), block.View).
			Absolute().Left(r.rect.X).Top(r.rect.Y).Width(r.rect.W).
			Height(block.Height).Radius(8)
	}
}
