package web

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/emiliopalmerini/quintaedizione.online/internal/content/catalog"
	"github.com/emiliopalmerini/quintaedizione.online/internal/content/release"
	"github.com/emiliopalmerini/quintaedizione.online/internal/content/routing"
)

const (
	testSpellID      = "b831eed4-f23a-4b3a-831b-f8c3dfd3831d"
	testCanonicalURL = "/compendio/incantesimi/5e/comunione"
	testLegacyURL    = "/srd/incantesimi/5e/comunione"
)

func TestHandlerRendersCanonicalEntityPage(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, testCanonicalURL, nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q", contentType)
	}
	body := response.Body.String()
	for _, expected := range []string{
		"<h1>Comunione</h1>",
		"Edizione 5e",
		`<link rel="canonical" href="/compendio/incantesimi/5e/comunione">`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body does not contain %q:\n%s", expected, body)
		}
	}
}

func TestHandlerRedirectsHistoricalAlias(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, testLegacyURL, nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want 301", response.Code)
	}
	if location := response.Header().Get("Location"); location != testCanonicalURL {
		t.Fatalf("Location = %q, want %q", location, testCanonicalURL)
	}
}

func TestHandlerRejectsEncodedPathVariant(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/compendio/incantesimi/5e/comunion%65", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestHandlerHeadMatchesGetMetadataWithoutBody(t *testing.T) {
	handler := testHandler(t)
	getResponse := httptest.NewRecorder()
	handler.ServeHTTP(getResponse, httptest.NewRequest(http.MethodGet, testCanonicalURL, nil))

	headResponse := httptest.NewRecorder()
	handler.ServeHTTP(headResponse, httptest.NewRequest(http.MethodHead, testCanonicalURL, nil))

	if headResponse.Code != http.StatusOK {
		t.Fatalf("HEAD status = %d, want 200", headResponse.Code)
	}
	if headResponse.Header().Get("Content-Length") != getResponse.Header().Get("Content-Length") {
		t.Fatalf("HEAD Content-Length = %q, GET = %q", headResponse.Header().Get("Content-Length"), getResponse.Header().Get("Content-Length"))
	}
	if headResponse.Body.Len() != 0 {
		t.Fatalf("HEAD body length = %d, want 0", headResponse.Body.Len())
	}
}

func TestHandlerReturnsNotFoundAndMethodNotAllowed(t *testing.T) {
	handler := testHandler(t)

	t.Run("unknown path", func(t *testing.T) {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/compendio/ignoto", nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", response.Code)
		}
	})

	t.Run("method", func(t *testing.T) {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, testCanonicalURL, nil))
		if response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", response.Code)
		}
		if allow := response.Header().Get("Allow"); allow != "GET, HEAD" {
			t.Fatalf("Allow = %q", allow)
		}
	})
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	resource := []byte(`[{
		"id":"` + testSpellID + `",
		"conceptId":"20f764dc-8db8-4aed-94f3-497a58f1e81d",
		"kind":"spell",
		"edition":"5e",
		"name":"Comunione",
		"revision":1,
		"source":{"scope":"srd","document":"srd-5e-it"}
	}]`)
	manifest := fmt.Sprintf(`{
		"formatVersion":1,
		"schemaVersion":"1.0.0",
		"datasetVersion":"2026.07.1",
		"resources":[{
			"path":"records.json",
			"recordKind":"spell",
			"mediaType":"application/json",
			"sha256":"%x"
		}]
	}`, sha256.Sum256(resource))
	loaded, err := release.Load(fstest.MapFS{
		"manifest.json": {Data: []byte(manifest)},
		"records.json":  {Data: resource},
	})
	if err != nil {
		t.Fatalf("release.Load() error = %v", err)
	}
	entities, err := catalog.Compile(loaded)
	if err != nil {
		t.Fatalf("catalog.Compile() error = %v", err)
	}
	routes, err := routing.New(entities, []routing.Entry{{
		EntityID: testSpellID,
		Path:     testCanonicalURL,
		Aliases:  []string{testLegacyURL},
	}})
	if err != nil {
		t.Fatalf("routing.New() error = %v", err)
	}
	handler, err := NewHandler(entities, routes)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return handler
}
