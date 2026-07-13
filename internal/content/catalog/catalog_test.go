package catalog

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/emiliopalmerini/quintaedizione.online/internal/content/release"
)

func TestCompileIndexesEntitiesAndConceptVersions(t *testing.T) {
	const (
		conceptID = "20f764dc-8db8-4aed-94f3-497a58f1e81d"
		spell5E   = "b831eed4-f23a-4b3a-831b-f8c3dfd3831d"
		spell55E  = "4d69974d-d6c6-4ae3-8e66-260fbdf354e2"
	)
	resource := []byte(fmt.Sprintf(`[
		{
			"id": %q,
			"conceptId": %q,
			"kind": "spell",
			"edition": "5e",
			"name": "Comunione",
			"revision": 1,
			"source": {"scope": "srd", "document": "srd-5e-it"},
			"level": 5
		},
		{
			"id": %q,
			"conceptId": %q,
			"kind": "spell",
			"edition": "5.5e",
			"name": "Comunione",
			"revision": 2,
			"source": {"scope": "srd", "document": "srd-5.5e-it"},
			"level": 5
		}
	]`, spell5E, conceptID, spell55E, conceptID))

	catalog, err := Compile(loadRelease(t, "spell", resource))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	got, exists := catalog.Entity(spell5E)
	if !exists {
		t.Fatalf("Entity(%q) was not found", spell5E)
	}
	if got.Name != "Comunione" || got.Edition != "5e" || got.Revision != 1 {
		t.Fatalf("Entity(%q) = %#v", spell5E, got)
	}

	versions := catalog.Versions(conceptID)
	if len(versions) != 2 {
		t.Fatalf("len(Versions()) = %d, want 2", len(versions))
	}
	if versions[0].Edition != "5.5e" || versions[1].Edition != "5e" {
		t.Fatalf("Versions() = %#v", versions)
	}
}

func TestCompileRejectsDuplicateEntityIdentity(t *testing.T) {
	resource := []byte(`[
		{
			"id":"b831eed4-f23a-4b3a-831b-f8c3dfd3831d",
			"conceptId":"20f764dc-8db8-4aed-94f3-497a58f1e81d",
			"kind":"spell",
			"edition":"5e",
			"name":"Comunione",
			"revision":1,
			"source":{"scope":"srd","document":"srd-5e-it"}
		},
		{
			"id":"b831eed4-f23a-4b3a-831b-f8c3dfd3831d",
			"conceptId":"6993f980-ce2d-44af-93a7-b71dfb25eca3",
			"kind":"spell",
			"edition":"5.5e",
			"name":"Comunione superiore",
			"revision":1,
			"source":{"scope":"srd","document":"srd-5.5e-it"}
		}
	]`)

	if _, err := Compile(loadRelease(t, "spell", resource)); err == nil {
		t.Fatal("Compile() succeeded with a duplicate entity ID")
	}
}

func TestCompileRejectsUninitializedRelease(t *testing.T) {
	if _, err := Compile(release.Release{}); err == nil {
		t.Fatal("Compile() succeeded with an uninitialized release")
	}
}

func TestCompileTracksValidTombstones(t *testing.T) {
	const retiredID = "b831eed4-f23a-4b3a-831b-f8c3dfd3831d"
	resource := []byte(`[{"id":"` + retiredID + `","reason":"removed from source"}]`)

	catalog, err := Compile(loadRelease(t, "tombstone", resource))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if !catalog.Retired(retiredID) {
		t.Fatalf("Retired(%q) = false", retiredID)
	}
}

func TestCompileRejectsInvalidTombstone(t *testing.T) {
	resource := []byte(`[{
		"id":"b831eed4-f23a-4b3a-831b-f8c3dfd3831d",
		"reason":""
	}]`)

	if _, err := Compile(loadRelease(t, "tombstone", resource)); err == nil {
		t.Fatal("Compile() succeeded with an invalid tombstone")
	}
}

func TestCompileRejectsInvalidIdentityAndUnsupportedEdition(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		edition string
	}{
		{name: "identity", id: "5e:spell:comunione", edition: "5e"},
		{name: "edition", id: "b831eed4-f23a-4b3a-831b-f8c3dfd3831d", edition: "6e"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource := []byte(fmt.Sprintf(`[{
				"id":%q,
				"conceptId":"20f764dc-8db8-4aed-94f3-497a58f1e81d",
				"kind":"spell",
				"edition":%q,
				"name":"Comunione",
				"revision":1,
				"source":{"scope":"srd","document":"srd-5e-it"}
			}]`, test.id, test.edition))

			if _, err := Compile(loadRelease(t, "spell", resource)); err == nil {
				t.Fatal("Compile() succeeded with invalid canonical metadata")
			}
		})
	}
}

func loadRelease(t *testing.T, recordKind string, resource []byte) release.Release {
	t.Helper()
	manifest := fmt.Sprintf(`{
		"formatVersion":1,
		"schemaVersion":"1.0.0",
		"datasetVersion":"2026.07.1",
		"resources":[{
			"path":"records.json",
			"recordKind":%q,
			"mediaType":"application/json",
			"sha256":"%x"
		}]
	}`, recordKind, sha256.Sum256(resource))
	loaded, err := release.Load(fstest.MapFS{
		"manifest.json": {Data: []byte(manifest)},
		"records.json":  {Data: resource},
	})
	if err != nil {
		t.Fatalf("release.Load() error = %v", err)
	}
	return loaded
}
