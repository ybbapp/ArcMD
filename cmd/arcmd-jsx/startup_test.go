package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const startupFixture = `{"version":1,"namespace":"00000000000000000000000000000000","markdown":"# Reviewed bundled demo","blocks":[]}`

func fixtureBundle(t *testing.T) (string, string, string) {
	t.Helper()
	app := filepath.Join(t.TempDir(), "Folder with spaces", "ArcMD JSX Preview.app", "Contents")
	if err := os.MkdirAll(filepath.Join(app, "MacOS"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(app, "Resources"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(app, "Resources", "demo.manifest.json")
	if err := os.WriteFile(path, []byte(startupFixture), 0644); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(startupFixture))
	return filepath.Join(app, "MacOS", "ArcMD-JSX-Preview"), path, hex.EncodeToString(digest[:])
}

func TestBundledDemoStartsWithoutCWDOrExternalTrust(t *testing.T) {
	executable, _, digest := fixtureBundle(t)
	for _, args := range [][]string{nil, {"-demo"}} {
		m, err := startupManifest(args, executable, digest, io.Discard)
		if err != nil || m.Markdown != "# Reviewed bundled demo" {
			t.Fatal(m, err)
		}
	}
}

func TestBundledDemoRejectsMissingModifiedOrUnpinnedResources(t *testing.T) {
	executable, path, digest := fixtureBundle(t)
	for _, expected := range []string{"", "wrong", strings.Repeat("0", 64)} {
		if _, err := startupManifest(nil, executable, expected, io.Discard); err == nil {
			t.Fatal("accepted unpinned or mismatched demo")
		}
	}
	if err := os.WriteFile(path, []byte(startupFixture+" "), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := startupManifest(nil, executable, digest, io.Discard); err == nil {
		t.Fatal("accepted modified demo")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := startupManifest(nil, executable, digest, io.Discard); err == nil {
		t.Fatal("accepted missing demo")
	}
}

func TestExternalManifestStillRequiresExplicitTrust(t *testing.T) {
	executable, path, digest := fixtureBundle(t)
	for _, args := range [][]string{{path}, {"-trust"}, {"-demo", path}, {"-demo", "-trust"}, {"-trust", path, path}} {
		if _, err := startupManifest(args, executable, digest, io.Discard); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if _, err := startupManifest([]string{"-trust", path}, executable, "", io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := startupManifest([]string{"-h"}, executable, "", io.Discard); !errors.Is(err, errShowHelp) {
		t.Fatal(err)
	}
}
