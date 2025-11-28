package mappers

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/display"
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/dto"
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/models"
)

// DocumentMapper handles conversion from raw documents to DTOs and Models
type DocumentMapper interface {
	ToDTO(collection string, item map[string]any) dto.DocumentDTO
	ToDTOs(collection string, items []map[string]any) []dto.DocumentDTO
	ToModel(collection string, item map[string]any) models.Document
	ToModels(collection string, items []map[string]any) []models.Document
}

type documentMapper struct {
	displayFactory *display.DisplayElementFactory
}

// NewDocumentMapper creates a new document mapper with display element factory
func NewDocumentMapper(displayFactory *display.DisplayElementFactory) DocumentMapper {
	return &documentMapper{
		displayFactory: displayFactory,
	}
}

// ToDTO converts a single raw document to DTO
func (m *documentMapper) ToDTO(collection string, item map[string]any) dto.DocumentDTO {
	dto := dto.DocumentDTO{}

	// Extract _id (slug) from document root
	if id, ok := item["_id"].(string); ok {
		dto.ID = id
	}

	// Extract title from Document model
	if title, ok := item["title"].(string); ok {
		dto.Title = title
	}

	// Extract translated flag from document root
	if translated, ok := item["translated"].(bool); ok {
		dto.Translated = translated
	}

	// Get display elements using the factory
	dto.DisplayElements = m.displayFactory.GetDisplayElements(collection, item)

	return dto
}

// ToDTOs converts multiple raw documents to DTOs
func (m *documentMapper) ToDTOs(collection string, items []map[string]any) []dto.DocumentDTO {
	documents := make([]dto.DocumentDTO, 0, len(items))

	for _, item := range items {
		doc := m.ToDTO(collection, item)
		documents = append(documents, doc)
	}

	return documents
}

// ToModel converts a single raw document to the existing models.Document format
func (m *documentMapper) ToModel(collection string, item map[string]any) models.Document {
	model := models.Document{}

	// Extract _id (slug) from document root
	if id, ok := item["_id"].(string); ok {
		model.ID = id
	}

	// Extract title from Document model
	if title, ok := item["title"].(string); ok {
		model.Title = title
	}

	// Extract translated flag from document root
	if translated, ok := item["translated"].(bool); ok {
		model.Translated = translated
	}

	// Get display elements using the factory and convert to model format
	displayElements := m.displayFactory.GetDisplayElements(collection, item)
	for _, elem := range displayElements {
		model.DisplayElements = append(model.DisplayElements, models.DocumentDisplayField{
			Value: elem.Value,
		})
	}

	return model
}

// ToModels converts multiple raw documents to models.Document format
func (m *documentMapper) ToModels(collection string, items []map[string]any) []models.Document {
	documents := make([]models.Document, 0, len(items))

	for _, item := range items {
		doc := m.ToModel(collection, item)
		documents = append(documents, doc)
	}

	return documents
}
