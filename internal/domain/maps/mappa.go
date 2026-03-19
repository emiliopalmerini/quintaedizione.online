package maps

// Mappa represents a map with Italian-translated metadata.
type Mappa struct {
	Slug          string
	Nome          string
	NomeOriginale string
	Immagine      string
	Categoria     string
	Tag           []string
	Descrizione   string
	Autore        string
	Licenza       string
	URLOriginale  string
}

// SearchFilters holds all possible filter criteria for map search.
type SearchFilters struct {
	Query     string
	Categoria string
	Tag       string
}

// Repository defines the interface for accessing map data.
type Repository interface {
	FindAll() []Mappa
	FindBySlug(slug string) (Mappa, bool)
	Search(filters SearchFilters) []Mappa
	Categorie() []string
	Tags() []string
}

// GalleryData holds everything needed to render the map gallery page.
type GalleryData struct {
	Mappe     []Mappa
	Categorie []string
	Tags      []string
	Query     string
	Categoria string
	Tag       string
	Total     int
}
