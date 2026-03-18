package web

import (
	"net/http"
	"strconv"
)

type PaginationData struct {
	TotalPages int
	StartItem  int
	EndItem    int
	HasNext    bool
	HasPrev    bool
}

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

// PaginationParams holds parsed pagination query parameters
type PaginationParams struct {
	PageNum  int
	PageSize int
	Query    string
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
