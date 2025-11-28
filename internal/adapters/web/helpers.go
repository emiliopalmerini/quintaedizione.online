package web

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
