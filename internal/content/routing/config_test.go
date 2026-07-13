package routing

import (
	"testing"
	"testing/fstest"
)

func TestLoadRegistryBuildsValidatedRoutes(t *testing.T) {
	files := fstest.MapFS{"routes.json": {Data: []byte(`{
		"version": 1,
		"routes": [{
			"entityId": "` + spellID + `",
			"path": "/compendio/incantesimi/5e/comunione",
			"aliases": ["/srd/incantesimi/5e/comunione"]
		}]
	}`)}}

	registry, err := LoadRegistry(files, "routes.json", testCatalog(t))
	if err != nil {
		t.Fatalf("LoadRegistry() error = %v", err)
	}
	resolved, exists := registry.Resolve("/srd/incantesimi/5e/comunione")
	if !exists || !resolved.Redirect || resolved.EntityID != spellID {
		t.Fatalf("Resolve(alias) = %#v, %t", resolved, exists)
	}
}

func TestLoadRegistryRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "version",
			data: `{"version":2,"routes":[]}`,
		},
		{
			name: "unknown field",
			data: `{"version":1,"routes":[],"slugs":[]}`,
		},
		{
			name: "unknown route field",
			data: `{"version":1,"routes":[{"entityId":"` + spellID + `","path":"/comunione","name":"Comunione"}]}`,
		},
		{
			name: "trailing JSON",
			data: `{"version":1,"routes":[]} {}`,
		},
		{
			name: "missing routes",
			data: `{"version":1}`,
		},
		{
			name: "null routes",
			data: `{"version":1,"routes":null}`,
		},
		{
			name: "empty routes",
			data: `{"version":1,"routes":[]}`,
		},
		{
			name: "duplicate top-level field",
			data: `{"version":1,"version":1,"routes":[{"entityId":"` + spellID + `","path":"/one"}]}`,
		},
		{
			name: "duplicate route field",
			data: `{"version":1,"routes":[{"entityId":"` + spellID + `","path":"/one","path":"/two"}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files := fstest.MapFS{"routes.json": {Data: []byte(test.data)}}
			if _, err := LoadRegistry(files, "routes.json", testCatalog(t)); err == nil {
				t.Fatal("LoadRegistry() succeeded with invalid configuration")
			}
		})
	}
}
