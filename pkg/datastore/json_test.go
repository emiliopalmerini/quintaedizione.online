package datastore

import (
	"testing"
	"testing/fstest"
)

type testItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestLoadJSON_Success(t *testing.T) {
	fs := fstest.MapFS{
		"data.json": &fstest.MapFile{
			Data: []byte(`[{"name":"a","value":1},{"name":"b","value":2}]`),
		},
	}

	items, err := LoadJSON[testItem](fs, "data.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "a" || items[0].Value != 1 {
		t.Errorf("unexpected first item: %+v", items[0])
	}
}

func TestLoadJSON_FileNotFound(t *testing.T) {
	fs := fstest.MapFS{}

	_, err := LoadJSON[testItem](fs, "missing.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadJSON_InvalidJSON(t *testing.T) {
	fs := fstest.MapFS{
		"bad.json": &fstest.MapFile{Data: []byte(`not json`)},
	}

	_, err := LoadJSON[testItem](fs, "bad.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
