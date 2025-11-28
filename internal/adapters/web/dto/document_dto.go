package dto

type DocumentDTO struct {
	ID              string              `json:"id"`
	Title           string              `json:"title"`
	DisplayElements []DisplayElementDTO `json:"display_elements"`
	Translated      bool                `json:"translated"`
}

type DisplayElementDTO struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}
