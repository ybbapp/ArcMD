package jsx

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// MDXCompilerVersion is the exact compiler target, not a full-conformance claim.
const MDXCompilerVersion = "3.1.1"
const CompatibilityProfile = "arcmd-mdx3-phase1"

//go:embed assets/build-info.json
var compilerBuildJSON string

// CompilerInfo identifies the actual embedded build in diagnostic manifests.
type CompilerInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Profile     string `json:"profile"`
	BuildSHA256 string `json:"buildSha256"`
}

type buildMetadata struct {
	SchemaVersion int               `json:"schemaVersion"`
	Profile       string            `json:"profile"`
	Compiler      CompilerInfo      `json:"compiler"`
	Dependencies  map[string]string `json:"dependencies"`
	Inputs        map[string]string `json:"inputs"`
	Assets        map[string]string `json:"assets"`
}

func compilerBuildIdentity(raw string) (CompilerInfo, error) {
	var meta buildMetadata
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return CompilerInfo{}, fmt.Errorf("compiler build metadata: %w", err)
	}
	if meta.SchemaVersion != 1 || meta.Profile != CompatibilityProfile || meta.Compiler.Name != "@mdx-js/mdx" || meta.Compiler.Version != MDXCompilerVersion || meta.Dependencies["@mdx-js/mdx"] != MDXCompilerVersion {
		return CompilerInfo{}, fmt.Errorf("compiler build does not match %s / MDX %s", CompatibilityProfile, MDXCompilerVersion)
	}
	for name, content := range map[string]string{
		"mdx-compiler.js":   compilerJS,
		"react-runtime.js":  reactRuntime,
		"island-wrapper.js": islandWrapper,
		"style.css":         islandStyle,
		"measure.js":        MeasureScript,
	} {
		if meta.Assets[name] != SourceSHA256(content) {
			return CompilerInfo{}, fmt.Errorf("embedded compiler asset hash mismatch: %s", name)
		}
	}
	return CompilerInfo{Name: meta.Compiler.Name, Version: meta.Compiler.Version, Profile: meta.Profile, BuildSHA256: SourceSHA256(raw)}, nil
}
