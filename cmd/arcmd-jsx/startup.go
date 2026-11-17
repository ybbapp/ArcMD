package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// The packaging script pins the reviewed demo artifact's digest at link time.
// A normal development binary has no default document and still requires -trust.
// No argument supplied by Finder is treated as consent to execute a user's file.
var bundledDemoSHA256 string

var errShowHelp = errors.New("help requested")

func startupManifest(args []string, executable, demoSHA256 string, output io.Writer) (manifest, error) {
	var empty manifest
	flags := flag.NewFlagSet("arcmd-jsx", flag.ContinueOnError)
	flags.SetOutput(output)
	trust := flags.Bool("trust", false, "execute only a locally compiled, reviewed manifest")
	demo := flags.Bool("demo", false, "open only the checksum-pinned bundled preview demo")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return empty, errShowHelp
		}
		return empty, err
	}
	if *demo || len(args) == 0 {
		if *trust || flags.NArg() != 0 {
			return empty, errors.New("-demo cannot be combined with -trust or file paths")
		}
		return bundledManifest(executable, demoSHA256)
	}
	if !*trust || flags.NArg() != 1 {
		return empty, errors.New("usage: arcmd-jsx -trust reviewed.manifest.json; the preview app opens only its bundled demo without arguments")
	}
	file, err := os.Open(flags.Arg(0))
	if err != nil {
		return empty, err
	}
	defer file.Close()
	return readManifest(file)
}

func bundledManifest(executable, expectedSHA256 string) (manifest, error) {
	var empty manifest
	digest, err := hex.DecodeString(expectedSHA256)
	if err != nil || len(digest) != sha256.Size {
		return empty, errors.New("no pinned bundled demo: use -trust with a reviewed manifest, or the packaged preview app")
	}
	// Resolve from the executable, never the working directory: Finder launch,
	// app translocation and paths containing spaces must not select another file.
	path := filepath.Join(filepath.Dir(executable), "..", "Resources", "demo.manifest.json")
	file, err := os.Open(path)
	if err != nil {
		return empty, fmt.Errorf("bundled demo: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, (16<<20)+1))
	if err != nil {
		return empty, err
	}
	if len(data) > 16<<20 {
		return empty, errors.New("bundled demo exceeds 16 MiB")
	}
	actual := sha256.Sum256(data)
	if !bytes.Equal(actual[:], digest) {
		return empty, errors.New("bundled demo checksum mismatch; refusing to execute modified content")
	}
	return readManifest(bytes.NewReader(data))
}
