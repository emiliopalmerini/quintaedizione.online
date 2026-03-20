package web

import (
	"net/http/httptest"
	"testing"
)

func TestExtractPaginationParams_DefaultValues(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	params := ExtractPaginationParams(r)

	if params.PageNum != 1 {
		t.Errorf("Expected default PageNum=1, got %d", params.PageNum)
	}
	if params.PageSize != 20 {
		t.Errorf("Expected default PageSize=20, got %d", params.PageSize)
	}
	if params.Query != "" {
		t.Errorf("Expected empty Query, got %s", params.Query)
	}
}

func TestExtractPaginationParams_ValidValues(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?page=3&page_size=50&q=dragon", nil)
	params := ExtractPaginationParams(r)

	if params.PageNum != 3 {
		t.Errorf("Expected PageNum=3, got %d", params.PageNum)
	}
	if params.PageSize != 50 {
		t.Errorf("Expected PageSize=50, got %d", params.PageSize)
	}
	if params.Query != "dragon" {
		t.Errorf("Expected Query='dragon', got '%s'", params.Query)
	}
}

func TestExtractPaginationParams_InvalidPage(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantPage int
	}{
		{"negative page", "/test?page=-1", 1},
		{"zero page", "/test?page=0", 1},
		{"non-numeric page", "/test?page=abc", 1},
		{"empty page", "/test?page=", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)
			params := ExtractPaginationParams(r)
			if params.PageNum != tt.wantPage {
				t.Errorf("Expected PageNum=%d, got %d", tt.wantPage, params.PageNum)
			}
		})
	}
}

func TestExtractPaginationParams_InvalidPageSize(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantSize int
	}{
		{"negative page_size", "/test?page_size=-1", 20},
		{"zero page_size", "/test?page_size=0", 20},
		{"too large page_size", "/test?page_size=101", 20},
		{"non-numeric page_size", "/test?page_size=abc", 20},
		{"empty page_size", "/test?page_size=", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)
			params := ExtractPaginationParams(r)
			if params.PageSize != tt.wantSize {
				t.Errorf("Expected PageSize=%d, got %d", tt.wantSize, params.PageSize)
			}
		})
	}
}

func TestParseIntParam(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		key        string
		defaultVal int
		want       int
	}{
		{"present valid", "/test?max_xp=500", "max_xp", 1000, 500},
		{"missing", "/test", "max_xp", 1000, 1000},
		{"empty", "/test?max_xp=", "max_xp", 1000, 1000},
		{"invalid", "/test?max_xp=abc", "max_xp", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)
			got := ParseIntParam(r, tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("ParseIntParam(%q, %d) = %d, want %d", tt.key, tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty", "", nil},
		{"single", "dungeon", []string{"dungeon"}},
		{"multiple", "dungeon,forest,cave", []string{"dungeon", "forest", "cave"}},
		{"with spaces", " dungeon , forest ", []string{"dungeon", "forest"}},
		{"trailing comma", "dungeon,", []string{"dungeon"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTags(tt.raw)
			if len(got) != len(tt.want) {
				t.Errorf("ParseTags(%q) = %v, want %v", tt.raw, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseTags(%q)[%d] = %q, want %q", tt.raw, i, got[i], tt.want[i])
				}
			}
		})
	}
}
