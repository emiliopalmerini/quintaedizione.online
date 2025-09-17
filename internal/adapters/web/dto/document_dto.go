package dto

// DocumentDTO represents a document for web presentation
type DocumentDTO struct {
	ID              string                `json:"id"`
	Nome            string                `json:"nome"`
	Slug            string                `json:"slug"`
	DisplayElements []DisplayElementDTO   `json:"display_elements"`
	Translated      bool                  `json:"translated"`
}

// DisplayElementDTO represents a display element for the UI
type DisplayElementDTO struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}