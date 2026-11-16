package main

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestIslandThemeLightAndDarkUseDifferentCSSVariables(t *testing.T) {
	light := newIslandTheme(ui.LightTheme())
	dark := newIslandTheme(ui.DarkTheme())
	if light.Dark || !dark.Dark {
		t.Fatalf("color-scheme: light.Dark=%v dark.Dark=%v", light.Dark, dark.Dark)
	}
	for _, name := range []string{"--bg", "--text", "--text-muted", "--surface", "--border", "--accent", "--accent-text", "--danger", "--font", "--font-size"} {
		if light.Vars[name] == "" || dark.Vars[name] == "" {
			t.Fatalf("missing %s", name)
		}
	}
	if light.Vars["--bg"] == dark.Vars["--bg"] || light.Vars["--text"] == dark.Vars["--text"] {
		t.Fatalf("light and dark paper/text should differ: %v vs %v", light.Vars, dark.Vars)
	}
	if light.Background() != light.Vars["--bg"] {
		t.Fatal("background token should match paper")
	}
	if strings.Contains(light.Script(), "unsafe-inline") {
		t.Fatal("theme script must not mention CSP relaxations")
	}
	if !strings.Contains(dark.Script(), `"dark":true`) || !strings.Contains(dark.Script(), dark.Vars["--bg"]) {
		t.Fatalf("script missing dark tokens: %s", dark.Script())
	}
	if light.Script() == dark.Script() {
		t.Fatal("light and dark scripts should differ")
	}
}

func TestIslandThemeUsesPagePaperColorsNotHardcodedWhite(t *testing.T) {
	custom := ui.DarkTheme()
	custom.Background = ui.Hex("#112233")
	custom.Text = ui.Hex("#abcdef")
	theme := newIslandTheme(custom)
	if theme.Vars["--bg"] != "#112233" || theme.Vars["--text"] != "#abcdef" {
		t.Fatalf("expected custom paper/text, got %v", theme.Vars)
	}
	if !strings.Contains(theme.Script(), "#112233") || strings.Contains(theme.Script(), `"--bg":"#ffffff"`) {
		t.Fatalf("script should carry the custom paper color: %s", theme.Script())
	}
}

func TestIslandThemeScriptIsSelfContainedJSON(t *testing.T) {
	theme := newIslandTheme(ui.LightTheme())
	theme.Vars["--font"] = `system-ui, "Noto Sans"`
	script := theme.Script()
	if strings.Contains(script, "\n") && !strings.Contains(script, `\n`) {
		// Newlines inside the JSON string would be escaped; a raw newline in
		// the surrounding JS could break Eval.
	}
	if !strings.Contains(script, `system-ui, \"Noto Sans\"`) && !strings.Contains(script, `system-ui, \u0022Noto Sans\u0022`) {
		t.Fatalf("font family was not JSON-escaped: %s", script)
	}
}
