package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestExtractPaginationParams_DefaultValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	params := ExtractPaginationParams(c)

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
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?page=3&page_size=50&q=dragon", nil)

	params := ExtractPaginationParams(c)

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

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", tt.url, nil)

			params := ExtractPaginationParams(c)

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

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", tt.url, nil)

			params := ExtractPaginationParams(c)

			if params.PageSize != tt.wantSize {
				t.Errorf("Expected PageSize=%d, got %d", tt.wantSize, params.PageSize)
			}
		})
	}
}

func TestExtractPaginationParams_MaxPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?page_size=100", nil)

	params := ExtractPaginationParams(c)

	if params.PageSize != 100 {
		t.Errorf("Expected PageSize=100 (max allowed), got %d", params.PageSize)
	}
}

func TestExtractPaginationParams_QueryWithSpaces(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?q=fire+spell", nil)

	params := ExtractPaginationParams(c)

	if params.Query != "fire spell" {
		t.Errorf("Expected Query='fire spell', got '%s'", params.Query)
	}
}
