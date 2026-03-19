package collections

import (
	"testing"
)

func TestGetGroups_ReturnsAllGroups(t *testing.T) {
	groups := GetGroups()

	expectedLabels := []string{"Personaggi", "Magia & Mostri", "Equipaggiamento", "Riferimento"}

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

func TestGetGroups_MagiaMostriContents(t *testing.T) {
	groups := GetGroups()
	magia := groups[1]

	expected := []CollectionName{Incantesimi, Mostri}
	if len(magia.Collections) != len(expected) {
		t.Fatalf("Magia & Mostri: expected %d collections, got %d", len(expected), len(magia.Collections))
	}
	for i, col := range expected {
		if magia.Collections[i] != col {
			t.Errorf("Magia & Mostri[%d]: expected %s, got %s", i, col, magia.Collections[i])
		}
	}
}

func TestGetGroups_RiferimentoContents(t *testing.T) {
	groups := GetGroups()
	riferimento := groups[3]

	expected := []CollectionName{Regole, Servizi, Glossario}
	if len(riferimento.Collections) != len(expected) {
		t.Fatalf("Riferimento: expected %d collections, got %d", len(expected), len(riferimento.Collections))
	}
	for i, col := range expected {
		if riferimento.Collections[i] != col {
			t.Errorf("Riferimento[%d]: expected %s, got %s", i, col, riferimento.Collections[i])
		}
	}
}
