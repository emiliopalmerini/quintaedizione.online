package collections_test

import (
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
)

func TestGetQuickFilter_Incantesimi(t *testing.T) {
	qf, ok := collections.GetQuickFilter(collections.Incantesimi)
	if !ok {
		t.Fatal("expected quick filter for incantesimi")
	}
	if qf.FilterName != "livello" {
		t.Errorf("expected filter name 'livello', got %q", qf.FilterName)
	}
	if len(qf.Chips) != 10 {
		t.Errorf("expected 10 chips (trucchetto + levels 1-9), got %d", len(qf.Chips))
	}
	// First chip is Trucchetto mapping to value "0"
	if qf.Chips[0].Label != "Trucchetto" {
		t.Errorf("expected first chip label 'Trucchetto', got %q", qf.Chips[0].Label)
	}
	if len(qf.Chips[0].Values) != 1 || qf.Chips[0].Values[0] != "0" {
		t.Errorf("expected first chip values [0], got %v", qf.Chips[0].Values)
	}
	// Second chip is "1°" mapping to value "1"
	if qf.Chips[1].Label != "1°" {
		t.Errorf("expected second chip label '1°', got %q", qf.Chips[1].Label)
	}
}

func TestGetQuickFilter_Mostri(t *testing.T) {
	qf, ok := collections.GetQuickFilter(collections.Mostri)
	if !ok {
		t.Fatal("expected quick filter for mostri")
	}
	if qf.FilterName != "grado_sfida" {
		t.Errorf("expected filter name 'grado_sfida', got %q", qf.FilterName)
	}
	if len(qf.Chips) != 8 {
		t.Errorf("expected 8 CR range chips, got %d", len(qf.Chips))
	}
	// First chip groups CR 0 and 1/8 and 1/4
	first := qf.Chips[0]
	if first.Label != "0-¼" {
		t.Errorf("expected first chip label '0-¼', got %q", first.Label)
	}
	if len(first.Values) != 3 {
		t.Errorf("expected first chip to have 3 values, got %d", len(first.Values))
	}
}

func TestGetQuickFilter_Equipaggiamenti(t *testing.T) {
	qf, ok := collections.GetQuickFilter(collections.Equipaggiamenti)
	if !ok {
		t.Fatal("expected quick filter for equipaggiamenti")
	}
	if qf.FilterName != "categoria" {
		t.Errorf("expected filter name 'categoria', got %q", qf.FilterName)
	}
	if len(qf.Chips) == 0 {
		t.Error("expected at least one chip for equipaggiamenti")
	}
}

func TestGetQuickFilter_OggettiMagici(t *testing.T) {
	qf, ok := collections.GetQuickFilter(collections.OggettiMagici)
	if !ok {
		t.Fatal("expected quick filter for oggetti_magici")
	}
	if qf.FilterName != "rarita" {
		t.Errorf("expected filter name 'rarita', got %q", qf.FilterName)
	}
}

func TestGetQuickFilter_CollectionWithoutQuickFilter(t *testing.T) {
	_, ok := collections.GetQuickFilter(collections.Regole)
	if ok {
		t.Error("expected no quick filter for regole")
	}
}
