package release

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"testing/fstest"
)

func TestLoadVerifiesRelease(t *testing.T) {
	resource := []byte(`[]`)
	manifest := validManifest(resource)

	got, err := Load(fstest.MapFS{
		"manifest.json":        {Data: []byte(manifest)},
		"entities/spells.json": {Data: resource},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	metadata := got.Manifest()
	if metadata.FormatVersion != 1 {
		t.Fatalf("FormatVersion = %d, want 1", metadata.FormatVersion)
	}
	if metadata.SchemaVersion != "1.0.0" {
		t.Fatalf("SchemaVersion = %q, want 1.0.0", metadata.SchemaVersion)
	}
	if metadata.DatasetVersion != "2026.07.1" {
		t.Fatalf("DatasetVersion = %q, want 2026.07.1", metadata.DatasetVersion)
	}
	if len(metadata.Resources) != 1 || metadata.Resources[0].Path != "entities/spells.json" {
		t.Fatalf("Resources = %#v", metadata.Resources)
	}
	spellResources := got.ResourcesByKind("spell")
	if len(spellResources) != 1 || spellResources[0].Path != "entities/spells.json" {
		t.Fatalf("ResourcesByKind(spell) = %#v", spellResources)
	}
	if resources := got.ResourcesByKind("monster"); len(resources) != 0 {
		t.Fatalf("ResourcesByKind(monster) = %#v, want none", resources)
	}
	gotResource, err := got.ReadResource("entities/spells.json")
	if err != nil {
		t.Fatalf("ReadResource() error = %v", err)
	}
	if string(gotResource) != string(resource) {
		t.Fatalf("ReadResource() = %q, want %q", gotResource, resource)
	}
}

func TestReleaseManifestReturnsDefensiveCopy(t *testing.T) {
	resource := []byte(`[]`)
	loaded, err := Load(fstest.MapFS{
		"manifest.json":        {Data: []byte(validManifest(resource))},
		"entities/spells.json": {Data: resource},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	metadata := loaded.Manifest()
	metadata.Resources[0].RecordKind = "tombstone"
	if resources := loaded.ResourcesByKind("spell"); len(resources) != 1 {
		t.Fatalf("ResourcesByKind(spell) = %#v after mutating manifest copy", resources)
	}
}

func TestReleaseRejectsUndeclaredResource(t *testing.T) {
	resource := []byte(`[]`)
	release, err := Load(fstest.MapFS{
		"manifest.json":        {Data: []byte(validManifest(resource))},
		"entities/spells.json": {Data: resource},
		"private.json":         {Data: []byte(`{}`)},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, err := release.ReadResource("private.json"); err == nil {
		t.Fatal("ReadResource() succeeded for an undeclared resource")
	}
}

func TestLoadRejectsUnknownManifestField(t *testing.T) {
	manifest := `{
		"formatVersion": 1,
		"schemaVersion": "1.0.0",
		"datasetVersion": "2026.07.1",
		"resources": [],
		"websiteSlug": "comunione"
	}`

	if _, err := Load(fstest.MapFS{"manifest.json": {Data: []byte(manifest)}}); err == nil {
		t.Fatal("Load() succeeded with an unknown manifest field")
	}
}

func TestLoadRejectsDuplicateManifestField(t *testing.T) {
	manifest := `{
		"formatVersion": 1,
		"schemaVersion": "1.0.0",
		"schemaVersion": "2.0.0",
		"datasetVersion": "2026.07.1",
		"resources": []
	}`

	if _, err := Load(fstest.MapFS{"manifest.json": {Data: []byte(manifest)}}); err == nil {
		t.Fatal("Load() succeeded with a duplicate manifest field")
	}
}

func TestLoadRejectsResourceWithoutRecordKind(t *testing.T) {
	manifest := `{
		"formatVersion": 1,
		"schemaVersion": "1.0.0",
		"datasetVersion": "2026.07.1",
		"resources": [{
			"path":"entities/spells.json",
			"mediaType":"application/json",
			"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		}]
	}`

	if _, err := Load(fstest.MapFS{"manifest.json": {Data: []byte(manifest)}}); err == nil {
		t.Fatal("Load() succeeded without a resource record kind")
	}
}

func TestLoadRejectsUnsupportedVersions(t *testing.T) {
	tests := []struct {
		name          string
		formatVersion int
		schemaVersion string
	}{
		{name: "release format", formatVersion: 2, schemaVersion: "1.0.0"},
		{name: "schema major", formatVersion: 1, schemaVersion: "2.0.0"},
		{name: "invalid schema version", formatVersion: 1, schemaVersion: "v1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := fmt.Sprintf(`{
				"formatVersion": %d,
				"schemaVersion": %q,
				"datasetVersion": "2026.07.1",
				"resources": []
			}`, test.formatVersion, test.schemaVersion)

			if _, err := Load(fstest.MapFS{"manifest.json": {Data: []byte(manifest)}}); err == nil {
				t.Fatal("Load() succeeded with an unsupported version")
			}
		})
	}
}

func TestLoadRejectsInvalidDatasetVersion(t *testing.T) {
	manifest := `{
		"formatVersion": 1,
		"schemaVersion": "1.0.0",
		"datasetVersion": "latest",
		"resources": []
	}`

	if _, err := Load(fstest.MapFS{"manifest.json": {Data: []byte(manifest)}}); err == nil {
		t.Fatal("Load() succeeded with an invalid dataset version")
	}
}

func TestLoadRejectsUnsafeDuplicateAndUnsortedPaths(t *testing.T) {
	tests := []struct {
		name      string
		resources string
	}{
		{
			name: "unsafe",
			resources: `[{
				"path":"../spells.json",
				"recordKind":"spell",
				"mediaType":"application/json",
				"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			}]`,
		},
		{
			name: "root",
			resources: `[{
				"path":".",
				"recordKind":"spell",
				"mediaType":"application/json",
				"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			}]`,
		},
		{
			name: "duplicate",
			resources: `[{
				"path":"entities/spells.json",
				"recordKind":"spell",
				"mediaType":"application/json",
				"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			},{
				"path":"entities/spells.json",
				"recordKind":"spell",
				"mediaType":"application/json",
				"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			}]`,
		},
		{
			name: "unsorted",
			resources: `[{
				"path":"entities/spells.json",
				"recordKind":"spell",
				"mediaType":"application/json",
				"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			},{
				"path":"entities/classes.json",
				"recordKind":"class",
				"mediaType":"application/json",
				"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			}]`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := fmt.Sprintf(`{
				"formatVersion": 1,
				"schemaVersion": "1.0.0",
				"datasetVersion": "2026.07.1",
				"resources": %s
			}`, test.resources)

			if _, err := Load(fstest.MapFS{"manifest.json": {Data: []byte(manifest)}}); err == nil {
				t.Fatal("Load() succeeded with invalid resource paths")
			}
		})
	}
}

func TestLoadRejectsMissingAndCorruptResources(t *testing.T) {
	resource := []byte(`[]`)
	manifest := validManifest(resource)

	t.Run("missing", func(t *testing.T) {
		if _, err := Load(fstest.MapFS{"manifest.json": {Data: []byte(manifest)}}); err == nil {
			t.Fatal("Load() succeeded without a declared resource")
		}
	})

	t.Run("checksum", func(t *testing.T) {
		if _, err := Load(fstest.MapFS{
			"manifest.json":        {Data: []byte(manifest)},
			"entities/spells.json": {Data: []byte(`[{}]`)},
		}); err == nil {
			t.Fatal("Load() succeeded with a checksum mismatch")
		}
	})
}

func TestLoadRejectsInvalidResourceShape(t *testing.T) {
	for _, resource := range [][]byte{[]byte(`null`), []byte(`{}`), []byte(`[1]`)} {
		if _, err := Load(fstest.MapFS{
			"manifest.json":        {Data: []byte(validManifest(resource))},
			"entities/spells.json": {Data: resource},
		}); err == nil {
			t.Fatalf("Load() succeeded with resource %s", resource)
		}
	}
}

func validManifest(resource []byte) string {
	return fmt.Sprintf(`{
		"formatVersion": 1,
		"schemaVersion": "1.0.0",
		"datasetVersion": "2026.07.1",
		"resources": [{
			"path": "entities/spells.json",
			"recordKind": "spell",
			"mediaType": "application/json",
			"sha256": "%x"
		}]
	}`, sha256.Sum256(resource))
}
