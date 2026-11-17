package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

// islandTheme is the CSS custom-property set injected into each JSX WebView.
// Paper and body text follow the editor theme (and therefore the page), while
// controls still receive Surface, Border, Accent and Danger.
type islandTheme struct {
	Dark bool              `json:"dark"`
	Vars map[string]string `json:"vars"`
}

func cssColor(c ui.Color) string {
	if c.A == 0 {
		return "transparent"
	}
	if c.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%.3f)", c.R, c.G, c.B, float32(c.A)/255)
}

func newIslandTheme(t *ui.Theme) islandTheme {
	if t == nil {
		t = ui.LightTheme()
	}
	font := strings.TrimSpace(t.Font)
	if font == "" {
		font = "system-ui, sans-serif"
	}
	size := t.FontSize
	if size <= 0 {
		size = 13
	}
	// Prefer the theme colors the native canvas actually paints with: the
	// preview does not pass a custom Page, so e.page is derived from Theme
	// and Background/Text are the paper and body text.
	return islandTheme{
		Dark: t.Dark,
		Vars: map[string]string{
			"--bg":          cssColor(t.Background),
			"--text":        cssColor(t.Text),
			"--text-muted":  cssColor(t.TextMuted),
			"--surface":     cssColor(t.Surface),
			"--border":      cssColor(t.Border),
			"--accent":      cssColor(t.Accent),
			"--accent-text": cssColor(t.AccentText),
			"--danger":      cssColor(t.Danger),
			"--font":        font,
			"--font-size":   fmt.Sprintf("%.0fpx", size),
		},
	}
}

func (t islandTheme) Key() string {
	return t.Script()
}

func (t islandTheme) Background() string {
	if bg := t.Vars["--bg"]; bg != "" {
		return bg
	}
	return "transparent"
}

// Script is evaluated with Page.Eval, which is not subject to the page CSP.
// It only writes CSS custom properties, so React state in the same document
// is not remounted.
func (t islandTheme) Script() string {
	payload, err := json.Marshal(t)
	if err != nil {
		return "false"
	}
	return `(function(t){var r=document.documentElement;if(!r)return false;r.style.setProperty("color-scheme",t.dark?"dark":"light");if(t.vars){for(var k in t.vars){r.style.setProperty(k,t.vars[k]);}}return true})(` + string(payload) + `)`
}
