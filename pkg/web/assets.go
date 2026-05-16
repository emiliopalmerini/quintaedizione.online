package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	cssBundlePath    string
	cssBundleContent []byte
	criticalCSS      string
)

// LoadCSSAssets reads CSS files from dir and prepares two payloads:
//   - critical: concatenated and exposed via CriticalCSS for inlining in <head>
//   - bundle:   concatenated, hashed, exposed via CSSBundlePath / CSSBundleHandler
//
// Order is preserved. Critical files should not overlap with bundle files.
func LoadCSSAssets(dir string, critical, bundle []string) error {
	criticalBytes, err := concat(dir, critical)
	if err != nil {
		return fmt.Errorf("critical css: %w", err)
	}
	criticalCSS = string(criticalBytes)

	bundleBytes, err := concat(dir, bundle)
	if err != nil {
		return fmt.Errorf("bundle css: %w", err)
	}
	cssBundleContent = bundleBytes
	sum := sha256.Sum256(bundleBytes)
	cssBundlePath = "/static/css/bundle." + hex.EncodeToString(sum[:8]) + ".css"
	return nil
}

func concat(dir string, files []string) ([]byte, error) {
	var b strings.Builder
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return nil, err
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	return []byte(b.String()), nil
}

// CSSBundlePath returns the public URL of the hashed CSS bundle.
func CSSBundlePath() string { return cssBundlePath }

// CriticalCSS returns the inline-ready CSS string for the document head.
func CriticalCSS() string { return criticalCSS }

// CSSBundleHandler serves the bundle with immutable cache headers.
// The hashed filename makes it safe to cache for a year.
func CSSBundleHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Write(cssBundleContent)
	})
}
