package web

import (
	"testing"
)

func TestBuildQuickFilterData_Incantesimi(t *testing.T) {
	data := buildQuickFilterData("incantesimi", map[string]string{})
	if data == nil {
		t.Fatal("expected quick filter data for incantesimi")
	}
	if data.FilterName != "livello" {
		t.Errorf("expected filter name 'livello', got %q", data.FilterName)
	}
	if len(data.Chips) != 10 {
		t.Errorf("expected 10 chips, got %d", len(data.Chips))
	}
	for _, chip := range data.Chips {
		if chip.Active {
			t.Errorf("chip %q should not be active with no filters", chip.Label)
		}
	}
}

func TestBuildQuickFilterData_ActiveChip(t *testing.T) {
	data := buildQuickFilterData("incantesimi", map[string]string{"livello": "0"})
	if data == nil {
		t.Fatal("expected quick filter data")
	}
	if !data.Chips[0].Active {
		t.Error("Trucchetto chip should be active when livello=0")
	}
	if data.Chips[1].Active {
		t.Error("1° chip should not be active when livello=0")
	}
}

func TestBuildQuickFilterData_MultiValueChipActive(t *testing.T) {
	// Mostri CR range "0-¼" maps to values "0", "1/8", "1/4"
	// All three must be in the filter for the chip to be active
	data := buildQuickFilterData("mostri", map[string]string{"grado_sfida": "0,1/8,1/4"})
	if data == nil {
		t.Fatal("expected quick filter data for mostri")
	}
	if !data.Chips[0].Active {
		t.Error("0-¼ chip should be active when all its values are selected")
	}
}

func TestBuildQuickFilterData_MultiValueChipPartial(t *testing.T) {
	// Only 2 of 3 values selected — chip should NOT be active
	data := buildQuickFilterData("mostri", map[string]string{"grado_sfida": "0,1/8"})
	if data == nil {
		t.Fatal("expected quick filter data for mostri")
	}
	if data.Chips[0].Active {
		t.Error("0-¼ chip should not be active when only 2 of 3 values selected")
	}
}

func TestBuildQuickFilterData_UnknownCollection(t *testing.T) {
	data := buildQuickFilterData("nonexistent", map[string]string{})
	if data != nil {
		t.Error("expected nil for unknown collection")
	}
}

func TestBuildQuickFilterData_CollectionWithoutQuickFilter(t *testing.T) {
	data := buildQuickFilterData("regole", map[string]string{})
	if data != nil {
		t.Error("expected nil for collection without quick filter")
	}
}

func TestChipMatchesFilter(t *testing.T) {
	tests := []struct {
		name         string
		chipValues   []string
		activeValues map[string]bool
		want         bool
	}{
		{"all match", []string{"a", "b"}, map[string]bool{"a": true, "b": true, "c": true}, true},
		{"partial match", []string{"a", "b"}, map[string]bool{"a": true}, false},
		{"no match", []string{"a"}, map[string]bool{"b": true}, false},
		{"single match", []string{"a"}, map[string]bool{"a": true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chipMatchesFilter(tt.chipValues, tt.activeValues)
			if got != tt.want {
				t.Errorf("chipMatchesFilter(%v, %v) = %v, want %v", tt.chipValues, tt.activeValues, got, tt.want)
			}
		})
	}
}
