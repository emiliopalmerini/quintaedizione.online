package content

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestNewApplicationComposesCanonicalContent(t *testing.T) {
	const (
		entityID     = "b831eed4-f23a-4b3a-831b-f8c3dfd3831d"
		canonicalURL = "/compendio/incantesimi/5e/comunione"
	)
	resource := []byte(`[{
		"id":"` + entityID + `",
		"conceptId":"20f764dc-8db8-4aed-94f3-497a58f1e81d",
		"kind":"spell",
		"edition":"5e",
		"name":"Comunione",
		"revision":1,
		"source":{"scope":"srd","document":"srd-5e-it"}
	}]`)
	releaseFiles := fstest.MapFS{
		"manifest.json": {Data: []byte(fmt.Sprintf(`{
			"formatVersion":1,
			"schemaVersion":"1.0.0",
			"datasetVersion":"2026.07.1",
			"resources":[{
				"path":"records.json",
				"recordKind":"spell",
				"mediaType":"application/json",
				"sha256":"%x"
			}]
		}`, sha256.Sum256(resource)))},
		"records.json": {Data: resource},
	}
	routeFiles := fstest.MapFS{"routes.json": {Data: []byte(`{
		"version":1,
		"routes":[{
			"entityId":"` + entityID + `",
			"path":"` + canonicalURL + `"
		}]
	}`)}}

	application, err := NewApplication(releaseFiles, routeFiles, "routes.json")
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if application.DatasetVersion() != "2026.07.1" {
		t.Fatalf("DatasetVersion() = %q", application.DatasetVersion())
	}

	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, canonicalURL, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
}

func TestNewApplicationPropagatesBoundaryErrors(t *testing.T) {
	if _, err := NewApplication(fstest.MapFS{}, fstest.MapFS{}, "routes.json"); err == nil {
		t.Fatal("NewApplication() succeeded without a canonical release")
	}
}
