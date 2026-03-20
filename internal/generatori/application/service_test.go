package application_test

import (
	"testing"
	"testing/fstest"

	"github.com/emiliopalmerini/quintaedizione.online/internal/generatori/application"
)

func singleColumnFS() fstest.MapFS {
	return fstest.MapFS{
		"test.json": &fstest.MapFile{
			Data: []byte(`{
				"id": "test",
				"name": "Test Table",
				"description": "A test table.",
				"die": "1d4",
				"items": ["Alpha", "Beta", "Gamma", "Delta"]
			}`),
		},
	}
}

func multiColumnFS() fstest.MapFS {
	return fstest.MapFS{
		"multi.json": &fstest.MapFile{
			Data: []byte(`{
				"id": "multi",
				"name": "Multi Table",
				"description": "A multi-column table.",
				"die": "1d3",
				"columns": [
					{"name": "Color", "items": ["Red", "Green", "Blue"]},
					{"name": "Shape", "items": ["Circle", "Square"]}
				]
			}`),
		},
	}
}

func TestNewService_LoadsSingleColumn(t *testing.T) {
	svc, err := application.NewService(singleColumnFS())
	if err != nil {
		t.Fatalf("NewService() error: %v", err)
	}

	tables := svc.List()
	if len(tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tables))
	}
	if tables[0].ID != "test" {
		t.Errorf("expected ID 'test', got %q", tables[0].ID)
	}
	if len(tables[0].Items) != 4 {
		t.Errorf("expected 4 items, got %d", len(tables[0].Items))
	}
}

func TestNewService_LoadsMultiColumn(t *testing.T) {
	svc, err := application.NewService(multiColumnFS())
	if err != nil {
		t.Fatalf("NewService() error: %v", err)
	}

	table, ok := svc.Get("multi")
	if !ok {
		t.Fatal("expected to find table 'multi'")
	}
	if !table.IsMultiColumn() {
		t.Fatal("expected multi-column table")
	}
	if len(table.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(table.Columns))
	}
}

func TestNewService_RejectsEmptyItems(t *testing.T) {
	fs := fstest.MapFS{
		"empty.json": &fstest.MapFile{
			Data: []byte(`{"id":"empty","name":"Empty","description":"No items","die":"1d0","items":[]}`),
		},
	}
	_, err := application.NewService(fs)
	if err == nil {
		t.Fatal("expected error for empty items, got nil")
	}
}

func TestNewService_RejectsEmptyColumn(t *testing.T) {
	fs := fstest.MapFS{
		"bad.json": &fstest.MapFile{
			Data: []byte(`{"id":"bad","name":"Bad","description":"Empty col","die":"1d0","columns":[{"name":"A","items":[]}]}`),
		},
	}
	_, err := application.NewService(fs)
	if err == nil {
		t.Fatal("expected error for empty column items, got nil")
	}
}

func TestGet_Found(t *testing.T) {
	svc, _ := application.NewService(singleColumnFS())

	table, ok := svc.Get("test")
	if !ok {
		t.Fatal("expected to find table 'test'")
	}
	if table.Name != "Test Table" {
		t.Errorf("expected name 'Test Table', got %q", table.Name)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc, _ := application.NewService(singleColumnFS())

	_, ok := svc.Get("nonexistent")
	if ok {
		t.Fatal("expected not to find table 'nonexistent'")
	}
}

func TestRoll_SingleColumn(t *testing.T) {
	svc, _ := application.NewService(singleColumnFS())

	valid := map[string]bool{"Alpha": true, "Beta": true, "Gamma": true, "Delta": true}

	for i := 0; i < 20; i++ {
		result, err := svc.Roll("test")
		if err != nil {
			t.Fatalf("Roll() error: %v", err)
		}
		if len(result.Entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(result.Entries))
		}
		if !valid[result.Entries[0].Value] {
			t.Errorf("Roll() returned %q, not in table", result.Entries[0].Value)
		}
	}
}

func TestRoll_MultiColumn(t *testing.T) {
	svc, _ := application.NewService(multiColumnFS())

	colors := map[string]bool{"Red": true, "Green": true, "Blue": true}
	shapes := map[string]bool{"Circle": true, "Square": true}

	for i := 0; i < 20; i++ {
		result, err := svc.Roll("multi")
		if err != nil {
			t.Fatalf("Roll() error: %v", err)
		}
		if len(result.Entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(result.Entries))
		}
		if result.Entries[0].Column != "Color" {
			t.Errorf("expected column 'Color', got %q", result.Entries[0].Column)
		}
		if !colors[result.Entries[0].Value] {
			t.Errorf("unexpected color %q", result.Entries[0].Value)
		}
		if result.Entries[1].Column != "Shape" {
			t.Errorf("expected column 'Shape', got %q", result.Entries[1].Column)
		}
		if !shapes[result.Entries[1].Value] {
			t.Errorf("unexpected shape %q", result.Entries[1].Value)
		}
	}
}

func TestRoll_NotFound(t *testing.T) {
	svc, _ := application.NewService(singleColumnFS())

	_, err := svc.Roll("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent table, got nil")
	}
}
