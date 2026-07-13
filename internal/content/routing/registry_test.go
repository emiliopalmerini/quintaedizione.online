package routing

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/emiliopalmerini/quintaedizione.online/internal/content/catalog"
	"github.com/emiliopalmerini/quintaedizione.online/internal/content/release"
)

const spellID = "b831eed4-f23a-4b3a-831b-f8c3dfd3831d"

func TestRegistryResolvesCanonicalPathAndAliases(t *testing.T) {
	registry, err := New(testCatalog(t), []Entry{{
		EntityID: spellID,
		Path:     "/compendio/incantesimi/5e/comunione",
		Aliases:  []string{"/srd/incantesimi/5e/comunione"},
	}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	route, exists := registry.Route(spellID)
	if !exists || route.Path != "/compendio/incantesimi/5e/comunione" {
		t.Fatalf("Route(%q) = %#v, %t", spellID, route, exists)
	}

	resolved, exists := registry.Resolve("/compendio/incantesimi/5e/comunione")
	if !exists || resolved.EntityID != spellID || resolved.Redirect {
		t.Fatalf("Resolve(canonical) = %#v, %t", resolved, exists)
	}
	resolved, exists = registry.Resolve("/srd/incantesimi/5e/comunione")
	if !exists || resolved.EntityID != spellID || !resolved.Redirect || resolved.CanonicalPath != route.Path {
		t.Fatalf("Resolve(alias) = %#v, %t", resolved, exists)
	}
}

func TestRegistryReturnsDefensiveRouteCopies(t *testing.T) {
	registry, err := New(testCatalog(t), []Entry{{
		EntityID: spellID,
		Path:     "/compendio/incantesimi/5e/comunione",
		Aliases:  []string{"/srd/incantesimi/5e/comunione"},
	}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	route, _ := registry.Route(spellID)
	route.Aliases[0] = "/changed"
	route, _ = registry.Route(spellID)
	if route.Aliases[0] != "/srd/incantesimi/5e/comunione" {
		t.Fatalf("Route() exposed mutable aliases: %#v", route.Aliases)
	}
}

func TestRegistryRejectsUnknownOrDuplicateEntities(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
	}{
		{
			name: "unknown",
			entries: []Entry{{
				EntityID: "4d69974d-d6c6-4ae3-8e66-260fbdf354e2",
				Path:     "/compendio/incantesimi/5e/ignoto",
			}},
		},
		{
			name: "duplicate",
			entries: []Entry{
				{EntityID: spellID, Path: "/compendio/incantesimi/5e/comunione"},
				{EntityID: spellID, Path: "/compendio/incantesimi/5e/comunione-2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New(testCatalog(t), test.entries); err == nil {
				t.Fatal("New() succeeded with invalid entity mapping")
			}
		})
	}
}

func TestRegistryRejectsInvalidAndCollidingPaths(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
	}{
		{name: "relative", entry: Entry{EntityID: spellID, Path: "compendio/comunione"}},
		{name: "query", entry: Entry{EntityID: spellID, Path: "/compendio/comunione?edition=5e"}},
		{name: "traversal", entry: Entry{EntityID: spellID, Path: "/compendio/../admin"}},
		{name: "encoded traversal", entry: Entry{EntityID: spellID, Path: "/compendio/%2e%2e/admin"}},
		{name: "backslash authority", entry: Entry{EntityID: spellID, Path: `/\evil.example`}},
		{name: "duplicate alias", entry: Entry{EntityID: spellID, Path: "/compendio/comunione", Aliases: []string{"/old", "/old"}}},
		{name: "canonical alias", entry: Entry{EntityID: spellID, Path: "/compendio/comunione", Aliases: []string{"/compendio/comunione"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New(testCatalog(t), []Entry{test.entry}); err == nil {
				t.Fatal("New() succeeded with an invalid path")
			}
		})
	}
}

func testCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	resource := []byte(`[{
		"id":"` + spellID + `",
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
	compiled, err := catalog.Compile(loaded)
	if err != nil {
		t.Fatalf("catalog.Compile() error = %v", err)
	}
	return compiled
}
