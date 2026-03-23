package models

type PageData struct {
	Title       string
	Description string
	Collection  string
	DocTitle    string
	DocID       string
	QueryString string
	TotalItems  int64
}

type Collection struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
}

type CollectionGroup struct {
	Label       string
	Collections []Collection
}

type HomePageData struct {
	PageData
	Collections []Collection
	Groups      []CollectionGroup
	Total       int64
	Editions    int
}

type Document struct {
	ID              string                 `json:"_id"`
	Title           string                 `json:"title"`
	DisplayElements []DocumentDisplayField `json:"display_elements"`
	Translated      bool                   `json:"translated"`
}

type DocumentDisplayField struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

type FilterValueOption struct {
	Value string
	Count int64
}

type FilterOption struct {
	Name          string
	Label         string
	Values        []FilterValueOption
	CurrentValue  string
	CurrentValues []string
}

type QuickFilterChip struct {
	Label  string
	Values []string
	Active bool
}

type QuickFilterData struct {
	FilterName string
	Chips      []QuickFilterChip
}

type CollectionPageData struct {
	PageData
	Documents   []Document
	Filters     []FilterOption
	QuickFilter *QuickFilterData
	Query       string
	Page        int
	PageSize    int
	Total       int64
	TotalPages  int
	HasNext     bool
	HasPrev     bool
	StartItem   int
	EndItem     int
}

type ItemPageData struct {
	PageData
	BodyRaw         string
	BodyHTML        string // used for collections without stat-block templates
	PrevID          string
	NextID          string
	Position        int
	Total           int
	CollectionLabel string
	SourceShort     string // edition badge (e.g. "5e", "5.5e"); empty if single source

	// Stat-block view models — at most one is set based on collection type.
	// The template checks these before falling back to BodyHTML.
	Spell   *SpellStatBlock
	Monster *MonsterStatBlock
	Class   *ClassStatBlock
	Species *SpeciesStatBlock
}

type ErrorPageData struct {
	PageData
	ErrorTitle   string
	ErrorMessage string
	ErrorCode    int
}

type CollectionSearchResult struct {
	CollectionName  string
	CollectionLabel string
	Documents       []Document
	Total           int64
	HasMore         bool
}

type SearchPageData struct {
	PageData
	Query       string
	Results     []CollectionSearchResult
	Total       int64
	Collections []Collection
}
