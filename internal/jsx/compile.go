package jsx

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrNoNode     = errors.New("JSX compiler requires a local Node.js (node on PATH), which is not bundled with ArcMD")
	ErrNoCompiler = errors.New("JSX compiler script not found")
	ErrCompile    = errors.New("JSX compilation failed")
)

// Options control how a document is compiled. Zero values are discovered.
type Options struct {
	Node     string
	Compiler string
	CacheDir string
	Timeout  time.Duration
}

func (o Options) withDefaults() Options {
	if o.Node == "" {
		o.Node = os.Getenv("ARCMD_JSX_NODE")
	}
	if o.Node == "" {
		o.Node = "node"
	}
	if o.Compiler == "" {
		o.Compiler = os.Getenv("ARCMD_JSX_COMPILER")
	}
	if o.Compiler == "" {
		if p, err := FindCompiler(); err == nil {
			o.Compiler = p
		}
	}
	if o.CacheDir == "" {
		o.CacheDir = os.Getenv("ARCMD_JSX_CACHE")
	}
	if o.Timeout <= 0 {
		o.Timeout = 30 * time.Second
	}
	return o
}

// FindCompiler locates experiments/jsx-webview/compile.mjs from the
// environment, the executable, or a source tree walked from the working
// directory.
func FindCompiler() (string, error) {
	var candidates []string
	if p := os.Getenv("ARCMD_JSX_COMPILER"); p != "" {
		candidates = append(candidates, p)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "experiments", "jsx-webview", "compile.mjs"),
			filepath.Join(dir, "..", "Resources", "jsx-webview", "compile.mjs"),
			filepath.Join(dir, "Resources", "jsx-webview", "compile.mjs"),
		)
	}
	if wd, err := os.Getwd(); err == nil {
		for dir := wd; ; dir = filepath.Dir(dir) {
			candidates = append(candidates, filepath.Join(dir, "experiments", "jsx-webview", "compile.mjs"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", ErrNoCompiler
}

// Compile runs the official island compiler. The source string is never
// written to the user's document path. A matching cache entry is reused.
// Missing Node or a compile error is returned so the host can show source.
func Compile(source string, opt Options) (Manifest, error) {
	opt = opt.withDefaults()
	var empty Manifest
	if opt.Compiler == "" {
		return empty, ErrNoCompiler
	}
	if _, err := exec.LookPath(opt.Node); err != nil {
		return empty, fmt.Errorf("%w: %v", ErrNoNode, err)
	}
	sum := SourceSHA256(source)
	if opt.CacheDir != "" {
		if m, err := readCache(opt.CacheDir, sum); err == nil {
			return m, nil
		}
	}
	inDir, err := os.MkdirTemp("", "arcmd-jsx-in-")
	if err != nil {
		return empty, err
	}
	defer os.RemoveAll(inDir)
	in := filepath.Join(inDir, "source.mdx")
	out := filepath.Join(inDir, "manifest.json")
	if err := os.WriteFile(in, []byte(source), 0600); err != nil {
		return empty, err
	}
	cmd := exec.Command(opt.Node, opt.Compiler, in, out)
	cmd.Dir = filepath.Dir(opt.Compiler)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr
	if err := runTimeout(cmd, opt.Timeout); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return empty, fmt.Errorf("%w: %s", ErrCompile, msg)
	}
	file, err := os.Open(out)
	if err != nil {
		return empty, fmt.Errorf("%w: %v", ErrCompile, err)
	}
	defer file.Close()
	m, err := ParseManifest(file)
	if err != nil {
		return empty, err
	}
	if m.SourceSHA256 != "" && m.SourceSHA256 != sum {
		return empty, fmt.Errorf("%w: source checksum mismatch", ErrCompile)
	}
	if m.SourceSHA256 == "" {
		m.SourceSHA256 = sum
	}
	if opt.CacheDir != "" {
		_ = writeCache(opt.CacheDir, sum, out)
	}
	return m, nil
}

func readCache(dir, sum string) (Manifest, error) {
	var empty Manifest
	file, err := os.Open(filepath.Join(dir, sum+".json"))
	if err != nil {
		return empty, err
	}
	defer file.Close()
	m, err := ParseManifest(file)
	if err != nil {
		return empty, err
	}
	if m.SourceSHA256 != sum {
		return empty, errors.New("stale JSX cache")
	}
	return m, nil
}

func writeCache(dir, sum, manifestPath string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, sum+".json.tmp")
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, sum+".json"))
}

func runTimeout(cmd *exec.Cmd, d time.Duration) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-time.After(d):
		_ = cmd.Process.Kill()
		<-done
		return fmt.Errorf("compiler timed out after %s", d)
	}
}
