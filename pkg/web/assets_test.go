package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScriptAssetsUseContentHashedImmutablePaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.js")
	if err := os.WriteFile(path, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadScriptAssets(dir, []string{"main.js"}); err != nil {
		t.Fatal(err)
	}
	firstPath := ScriptAssetPath("main.js")
	if !strings.HasPrefix(firstPath, "/static/main.") || !strings.HasSuffix(firstPath, ".js") {
		t.Fatalf("unexpected path %q", firstPath)
	}

	rec := httptest.NewRecorder()
	ScriptAssetHandler("main.js").ServeHTTP(rec, httptest.NewRequest(http.MethodGet, firstPath, nil))
	if rec.Body.String() != "first" {
		t.Errorf("body = %q", rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
		t.Errorf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}

	if err := os.WriteFile(path, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadScriptAssets(dir, []string{"main.js"}); err != nil {
		t.Fatal(err)
	}
	if secondPath := ScriptAssetPath("main.js"); secondPath == firstPath {
		t.Fatal("script path did not change with content")
	}
}
