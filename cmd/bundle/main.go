// Command bundle builds the single-file app: it inlines the CSS, the
// JavaScript, Go's wasm_exec.js, and the gzipped WebAssembly engine into
// web/index.html, so dist/recipe-math.html works offline with no server.
//
//	go run ./cmd/bundle -wasm build/recipemath.wasm -out dist/recipe-math.html
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	wasmPath := flag.String("wasm", "build/recipemath.wasm", "compiled engine")
	webDir := flag.String("web", "web", "folder with index.html, style.css, app.js")
	out := flag.String("out", "dist/recipe-math.html", "output file")
	flag.Parse()
	if err := run(*wasmPath, *webDir, *out); err != nil {
		fmt.Fprintln(os.Stderr, "bundle:", err)
		os.Exit(1)
	}
}

func run(wasmPath, webDir, out string) error {
	read := func(path string) (string, error) {
		b, err := os.ReadFile(path)
		return string(b), err
	}
	page, err := read(filepath.Join(webDir, "index.html"))
	if err != nil {
		return err
	}
	css, err := read(filepath.Join(webDir, "style.css"))
	if err != nil {
		return err
	}
	app, err := read(filepath.Join(webDir, "app.js"))
	if err != nil {
		return err
	}
	root, err := goRoot()
	if err != nil {
		return err
	}
	goJS, err := read(filepath.Join(root, "lib", "wasm", "wasm_exec.js"))
	if err != nil {
		return err
	}
	wasm, err := os.ReadFile(wasmPath)
	if err != nil {
		return err
	}
	var gz bytes.Buffer
	zw, err := gzip.NewWriterLevel(&gz, gzip.BestCompression)
	if err != nil {
		return err
	}
	if _, err := zw.Write(wasm); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}

	for _, s := range []string{css, app, goJS} {
		if strings.Contains(strings.ToLower(s), "</script") || strings.Contains(strings.ToLower(s), "</style") {
			return fmt.Errorf("an inlined file contains a closing tag")
		}
	}
	replacements := []string{
		"/*STYLE*/", css,
		"/*WASM_EXEC*/", goJS,
		"/*APP*/", app,
		"__WASM_GZIP_BASE64__", base64.StdEncoding.EncodeToString(gz.Bytes()),
	}
	for i := 0; i < len(replacements); i += 2 {
		if strings.Count(page, replacements[i]) != 1 {
			return fmt.Errorf("index.html must contain %s exactly once", replacements[i])
		}
	}
	page = strings.NewReplacer(replacements...).Replace(page)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, []byte(page), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%.1f MB)\n", out, float64(len(page))/1e6)
	return nil
}

// goRoot asks the go tool where Go is installed, since wasm_exec.js must
// match the compiler that built the engine.
func goRoot() (string, error) {
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return "", fmt.Errorf("finding GOROOT: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
