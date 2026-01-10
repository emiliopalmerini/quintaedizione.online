package parsers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKeywordLinker_LoadKeywords(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "keywords.json")

	configContent := `{
		"damage_types": {
			"Contundenti": "/regole/tipi-di-danno",
			"Taglienti": "/regole/tipi-di-danno"
		},
		"conditions": {
			"Affascinato": "/regole/affascinato",
			"Prono": "/regole/prono"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	linker, err := NewKeywordLinker(configPath)
	if err != nil {
		t.Fatalf("NewKeywordLinker failed: %v", err)
	}

	if linker == nil {
		t.Fatal("Expected non-nil linker")
	}

	// Verify keywords were loaded
	expectedKeywords := 4
	if len(linker.keywords) != expectedKeywords {
		t.Errorf("Expected %d keywords, got %d", expectedKeywords, len(linker.keywords))
	}

	// Verify specific keywords
	if url, exists := linker.keywords["Contundenti"]; !exists {
		t.Error("Expected keyword 'Contundenti' to exist")
	} else if url != "/regole/tipi-di-danno" {
		t.Errorf("Expected URL '/regole/tipi-di-danno', got '%s'", url)
	}
}

func TestKeywordLinker_LinkKeywords_SimpleText(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "keywords.json")

	configContent := `{
		"damage_types": {
			"Contundenti": "/regole/tipi-di-danno"
		},
		"conditions": {
			"Affascinato": "/regole/affascinato"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	linker, err := NewKeywordLinker(configPath)
	if err != nil {
		t.Fatalf("NewKeywordLinker failed: %v", err)
	}

	input := "<p>I danni Contundenti sono comuni.</p>"
	result := linker.LinkKeywords(input)

	if !strings.Contains(result, `href="/regole/tipi-di-danno"`) {
		t.Error("Expected link to damage types")
	}

	if !strings.Contains(result, `class="keyword-link"`) {
		t.Error("Expected keyword-link class")
	}

	if !strings.Contains(result, ">Contundenti<") {
		t.Error("Expected keyword text to be preserved")
	}
}

func TestKeywordLinker_LinkKeywords_FirstOccurrenceOnly(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "keywords.json")

	configContent := `{
		"damage_types": {
			"Contundenti": "/regole/tipi-di-danno"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	linker, err := NewKeywordLinker(configPath)
	if err != nil {
		t.Fatalf("NewKeywordLinker failed: %v", err)
	}

	input := "<p>Contundenti first. Contundenti second.</p>"
	result := linker.LinkKeywords(input)

	// Should link ALL occurrences (changed behavior from first-only to all)
	linkCount := strings.Count(result, `href="/regole/tipi-di-danno"`)
	if linkCount != 2 {
		t.Errorf("Expected 2 links (all occurrences), got %d", linkCount)
	}
}

func TestKeywordLinker_LinkKeywords_SkipCodeBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "keywords.json")

	configContent := `{
		"damage_types": {
			"Contundenti": "/regole/tipi-di-danno"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	linker, err := NewKeywordLinker(configPath)
	if err != nil {
		t.Fatalf("NewKeywordLinker failed: %v", err)
	}

	input := "<pre><code>Contundenti in code</code></pre><p>Contundenti in text</p>"
	result := linker.LinkKeywords(input)

	// Should only have one link (from the paragraph, not the code block)
	linkCount := strings.Count(result, `href="/regole/tipi-di-danno"`)
	if linkCount != 1 {
		t.Errorf("Expected 1 link (skipping code block), got %d", linkCount)
	}
}

func TestKeywordLinker_LinkKeywords_SkipExistingLinks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "keywords.json")

	configContent := `{
		"damage_types": {
			"Contundenti": "/regole/tipi-di-danno"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	linker, err := NewKeywordLinker(configPath)
	if err != nil {
		t.Fatalf("NewKeywordLinker failed: %v", err)
	}

	input := `<p><a href="/other">Contundenti link</a> and Contundenti text</p>`
	result := linker.LinkKeywords(input)

	// Should have one keyword-link (from the text, not the existing link)
	keywordLinkCount := strings.Count(result, `class="keyword-link"`)
	if keywordLinkCount != 1 {
		t.Errorf("Expected 1 keyword-link, got %d", keywordLinkCount)
	}
}

func TestKeywordLinker_LinkKeywords_EmptyInput(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "keywords.json")

	configContent := `{"damage_types": {}}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	linker, err := NewKeywordLinker(configPath)
	if err != nil {
		t.Fatalf("NewKeywordLinker failed: %v", err)
	}

	result := linker.LinkKeywords("")
	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
}

func TestKeywordLinker_LinkKeywords_LongestMatchFirst(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "keywords.json")

	configContent := `{
		"conditions": {
			"Privo di Sensi": "/regole/privo-di-sensi",
			"Prono": "/regole/prono"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	linker, err := NewKeywordLinker(configPath)
	if err != nil {
		t.Fatalf("NewKeywordLinker failed: %v", err)
	}

	input := "<p>La creatura è Privo di Sensi.</p>"
	result := linker.LinkKeywords(input)

	// Should link to "Privo di Sensi", not just "Prono"
	if !strings.Contains(result, `href="/regole/privo-di-sensi"`) {
		t.Error("Expected link to 'Privo di Sensi'")
	}

	if strings.Contains(result, `href="/regole/prono"`) {
		t.Error("Should not link to 'Prono' when 'Privo di Sensi' matches")
	}
}
