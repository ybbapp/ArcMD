package jsx

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCompilerBuildPinsAndInputs(t *testing.T) {
	identity, err := compilerBuildIdentity(compilerBuildJSON)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Version != "3.1.1" || identity.Profile != CompatibilityProfile || identity.BuildSHA256 == "" {
		t.Fatalf("identity: %+v", identity)
	}
	var meta buildMetadata
	if err = json.Unmarshal([]byte(compilerBuildJSON), &meta); err != nil {
		t.Fatal(err)
	}
	for name, want := range meta.Inputs {
		data, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		if SourceSHA256(string(data)) != want {
			t.Fatalf("regenerate compiler metadata after changing %s", name)
		}
	}
	bad := strings.Replace(compilerBuildJSON, meta.Assets["react-runtime.js"], strings.Repeat("0", 64), 1)
	if _, err = compilerBuildIdentity(bad); err == nil {
		t.Fatal("accepted mixed/stale generated assets")
	}
}

func TestMDX3ProjectionRestoresNewLanguageForms(t *testing.T) {
	for _, source := range []string{
		"# Native\n\n{1+2}\n\nAfter.\n",
		"# Native\n\n中文 😀 with <span>{6*7}</span> inline.\n\nAfter.\n",
		"# Native\n\n<div>\n\n<Counter initial={5} />\n\n</div>\n\nAfter.\n",
	} {
		m, err := Compile(source, Options{})
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Blocks) != 1 || m.Blocks[0].HTML == "" {
			t.Fatalf("projection: %+v", m)
		}
		if RestoreSource(m.Markdown, m.Namespace) != source {
			t.Fatalf("source changed: %q", source)
		}
		if m.Compiler.Version != MDXCompilerVersion || m.SourceSHA256 != SourceSHA256(source) {
			t.Fatal("wrong compiler/source identity")
		}
	}
}
