package web

import (
	"net/http"
	"strconv"
	"strings"
)

// PaginationParams holds parsed pagination query parameters.
type PaginationParams struct {
	PageNum  int
	PageSize int
	Query    string
}

// PaginationData holds computed pagination display data.
type PaginationData struct {
	TotalPages int
	StartItem  int
	EndItem    int
	HasNext    bool
	HasPrev    bool
}

// ExtractPaginationParams extracts and validates pagination parameters from the request.
// Returns default values: page=1, pageSize=20 if invalid or missing.
func ExtractPaginationParams(r *http.Request) PaginationParams {
	q := r.URL.Query()

	page := q.Get("page")
	if page == "" {
		page = "1"
	}
	pageSize := q.Get("page_size")
	if pageSize == "" {
		pageSize = "20"
	}
	query := q.Get("q")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 || pageSizeNum > 100 {
		pageSizeNum = 20
	}

	return PaginationParams{
		PageNum:  pageNum,
		PageSize: pageSizeNum,
		Query:    query,
	}
}

// CalculatePaginationData computes display-oriented pagination values.
func CalculatePaginationData(pageNum, pageSize int, totalCount int64) *PaginationData {
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	startItem := (pageNum-1)*pageSize + 1
	endItem := pageNum * pageSize

	if endItem > int(totalCount) {
		endItem = int(totalCount)
	}

	return &PaginationData{
		TotalPages: totalPages,
		StartItem:  startItem,
		EndItem:    endItem,
		HasNext:    pageNum < totalPages,
		HasPrev:    pageNum > 1,
	}
}

// ParseIntParam parses an integer query parameter with a default value.
// Returns defaultVal when the parameter is missing, empty, or invalid.
func ParseIntParam(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

// ParseTags splits a comma-separated tag string into a cleaned slice.
// Returns nil for an empty input.
func ParseTags(raw string) []string {
	if raw == "" {
		return nil
	}
	var tags []string
	for _, t := range strings.Split(raw, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
