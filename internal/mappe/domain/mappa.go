package domain

// Mappa represents a map with Italian-translated metadata.
type Mappa struct {
	Slug                 string
	Nome                 string
	NomeOriginale        string
	Immagine             string
	Tag                  []string
	Descrizione          string
	Autore               string
	Licenza              string
	URLOriginale         string
	URLImmagineOriginale string
}

// SearchFilters holds all possible filter criteria for map search.
type SearchFilters struct {
	Query  string
	Tags   []string
	Offset int
	Limit  int // 0 means no limit
}

// Repository defines the interface for accessing map data.
type Repository interface {
	FindAll() []Mappa
	FindBySlug(slug string) (Mappa, bool)
	Search(filters SearchFilters) ([]Mappa, int)
	Tags() []string
}

// GalleryData holds everything needed to render the map gallery page.
type GalleryData struct {
	Mappe      []Mappa
	Tags       []string
	Query      string
	ActiveTags []string
	Total      int
	Offset     int
	Limit      int
	HasMore    bool
}
