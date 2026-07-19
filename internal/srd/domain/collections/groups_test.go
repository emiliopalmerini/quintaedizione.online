package collections

import (
	"testing"
)

func TestGetGroups_ReturnsAllGroups(t *testing.T) {
	groups := GetGroups()

	expectedLabels := []string{"Personaggi", "Regole", "Magia", "Equipaggiamento", "Bestiario"}

	if len(groups) != len(expectedLabels) {
		t.Fatalf("Expected %d groups, got %d", len(expectedLabels), len(groups))
	}

	for i, group := range groups {
		if group.Label != expectedLabels[i] {
			t.Errorf("Group %d: expected label %q, got %q", i, expectedLabels[i], group.Label)
		}
	}
}

func TestGetGroups_OrderIsStable(t *testing.T) {
	groups1 := GetGroups()
	groups2 := GetGroups()

	if len(groups1) != len(groups2) {
		t.Fatal("Group count differs between calls")
	}

	for i := range groups1 {
		if groups1[i].Label != groups2[i].Label {
			t.Errorf("Group %d label differs: %q vs %q", i, groups1[i].Label, groups2[i].Label)
		}
		if len(groups1[i].Collections) != len(groups2[i].Collections) {
			t.Errorf("Group %q collection count differs", groups1[i].Label)
		}
		for j := range groups1[i].Collections {
			if groups1[i].Collections[j] != groups2[i].Collections[j] {
				t.Errorf("Group %q collection %d differs: %s vs %s",
					groups1[i].Label, j, groups1[i].Collections[j], groups2[i].Collections[j])
			}
		}
	}
}

func TestGetGroups_CoversAllCollections(t *testing.T) {
	groups := GetGroups()

	seen := make(map[CollectionName]bool)
	for _, group := range groups {
		for _, col := range group.Collections {
			if seen[col] {
				t.Errorf("Collection %s appears in multiple groups", col)
			}
			seen[col] = true
		}
	}

	for name := range Registry {
		if !seen[name] {
			t.Errorf("Collection %s is not assigned to any group", name)
		}
	}
}

func TestGetGroups_AllCollectionsAreValid(t *testing.T) {
	groups := GetGroups()

	for _, group := range groups {
		for _, col := range group.Collections {
			if _, exists := Registry[col]; !exists {
				t.Errorf("Group %q references invalid collection %s", group.Label, col)
			}
		}
	}
}

func TestGetGroups_AllHaveSlugAndDescription(t *testing.T) {
	groups := GetGroups()

	seen := make(map[string]bool)
	for _, g := range groups {
		if g.Slug == "" {
			t.Errorf("Group %q is missing a slug", g.Label)
		}
		if g.Description == "" {
			t.Errorf("Group %q is missing a description", g.Label)
		}
		if seen[g.Slug] {
			t.Errorf("Slug %q appears in more than one group", g.Slug)
		}
		seen[g.Slug] = true
	}
}

func TestGetGroups_PersonaggiContents(t *testing.T) {
	groups := GetGroups()
	personaggi := groups[0]

	expected := []CollectionName{Classi, Specie, Backgrounds, Talenti}
	if len(personaggi.Collections) != len(expected) {
		t.Fatalf("Personaggi: expected %d collections, got %d", len(expected), len(personaggi.Collections))
	}
	for i, col := range expected {
		if personaggi.Collections[i] != col {
			t.Errorf("Personaggi[%d]: expected %s, got %s", i, col, personaggi.Collections[i])
		}
	}
}

func TestGetGroups_RegoleContents(t *testing.T) {
	groups := GetGroups()
	regole := groups[1]

	expected := []CollectionName{Regole, Glossario}
	if len(regole.Collections) != len(expected) {
		t.Fatalf("Regole: expected %d collections, got %d", len(expected), len(regole.Collections))
	}
	for i, col := range expected {
		if regole.Collections[i] != col {
			t.Errorf("Regole[%d]: expected %s, got %s", i, col, regole.Collections[i])
		}
	}
}

func TestGetGroups_EquipaggiamentoContents(t *testing.T) {
	groups := GetGroups()
	equipaggiamento := groups[3]

	expected := []CollectionName{Equipaggiamenti, OggettiMagici, Servizi}
	if len(equipaggiamento.Collections) != len(expected) {
		t.Fatalf("Equipaggiamento: expected %d collections, got %d", len(expected), len(equipaggiamento.Collections))
	}
	for i, col := range expected {
		if equipaggiamento.Collections[i] != col {
			t.Errorf("Equipaggiamento[%d]: expected %s, got %s", i, col, equipaggiamento.Collections[i])
		}
	}
}

func TestGetGroup_KnownSlug(t *testing.T) {
	g, ok := GetGroup("personaggi")
	if !ok {
		t.Fatal("GetGroup(\"personaggi\") returned ok=false")
	}
	if g.Label != "Personaggi" {
		t.Errorf("Expected label Personaggi, got %q", g.Label)
	}
}

func TestGetGroup_UnknownSlug(t *testing.T) {
	if _, ok := GetGroup("bogus"); ok {
		t.Error("GetGroup(\"bogus\") expected ok=false, got true")
	}
}
