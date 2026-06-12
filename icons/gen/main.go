// Command gen vendors Lucide SVG icons into the icons/lucide directory.
//
// Usage:
//
//	go run ./icons/gen [-out dir] <name> [<name> ...]
//
// Each name is a Lucide icon identifier such as "check", "x", or
// "chevron-down". The pristine SVG is downloaded from the Lucide
// repository (raw.githubusercontent.com) and written to
// icons/lucide/<name>.svg. Normalization for the gogpu/gg renderer
// happens at runtime in the icons package, so the vendored files stay
// byte-identical to upstream.
//
// After vendoring, add a matching exported variable to icons/icons.go
// and list it in all().
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

// rawURLFormat is the download URL template for pristine Lucide SVGs.
const rawURLFormat = "https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/%s.svg"

// maxSVGSize caps the accepted response size; Lucide icons are < 2 KB.
const maxSVGSize = 64 * 1024

// nameRe matches valid Lucide icon names: lowercase words separated by
// single hyphens, e.g. "check", "chevron-down", "loader-circle".
var nameRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func main() {
	out := flag.String("out", defaultOutDir(), "output directory for downloaded SVGs")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: go run ./icons/gen [-out dir] <name> [<name> ...]\n\n"+
				"Downloads Lucide icons (e.g. \"check\", \"x\", \"chevron-down\") into %s.\n\n",
			*out)
		flag.PrintDefaults()
	}
	flag.Parse()

	names := flag.Args()
	if len(names) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	failed := 0
	for _, name := range names {
		if err := vendorIcon(client, *out, name); err != nil {
			fmt.Fprintf(os.Stderr, "gen: %s: %v\n", name, err)
			failed++
			continue
		}
		fmt.Printf("vendored %s -> %s\n", name, filepath.Join(*out, name+".svg"))
	}
	if failed > 0 {
		os.Exit(1)
	}
}

// vendorIcon downloads one Lucide icon and writes it into outDir.
func vendorIcon(client *http.Client, outDir, name string) error {
	if !nameRe.MatchString(name) {
		return fmt.Errorf("invalid icon name (want lowercase-hyphenated, e.g. \"chevron-down\")")
	}

	url := fmt.Sprintf(rawURLFormat, name)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %s", url, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSVGSize))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if !looksLikeSVG(body) {
		return fmt.Errorf("response does not look like an SVG document")
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	path := filepath.Join(outDir, name+".svg")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// looksLikeSVG reports whether data starts with an <svg or <?xml prolog
// after optional leading whitespace.
func looksLikeSVG(data []byte) bool {
	i := 0
	for i < len(data) {
		switch data[i] {
		case ' ', '\t', '\r', '\n':
			i++
			continue
		}
		break
	}
	rest := data[i:]
	return hasPrefix(rest, "<svg") || hasPrefix(rest, "<?xml")
}

// hasPrefix is bytes.HasPrefix for a string prefix without extra imports.
func hasPrefix(data []byte, prefix string) bool {
	if len(data) < len(prefix) {
		return false
	}
	return string(data[:len(prefix)]) == prefix
}

// defaultOutDir resolves the icons/lucide directory relative to this
// source file, so the tool works regardless of the current working
// directory when invoked via "go run".
func defaultOutDir() string {
	if _, file, _, ok := runtime.Caller(0); ok {
		return filepath.Join(filepath.Dir(file), "..", "lucide")
	}
	return filepath.Join("icons", "lucide")
}
