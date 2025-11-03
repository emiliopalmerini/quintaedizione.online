package dto

// DocumentDTO represents a document for web presentation
type DocumentDTO struct {
	ID              string              `json:"id"` // Slug from Document._id
	Title           string              `json:"title"` // Display name from Document.title
	DisplayElements []DisplayElementDTO `json:"display_elements"`
	Translated      bool                `json:"translated"`
}

// DisplayElementDTO represents a display element for the UI
type DisplayElementDTO struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}
