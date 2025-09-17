package mappers

import (
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/display"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/dto"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/models"
)

// DocumentMapper handles conversion from raw documents to DTOs and Models
type DocumentMapper interface {
	ToDTO(collection string, item map[string]interface{}) dto.DocumentDTO
	ToDTOs(collection string, items []map[string]interface{}) []dto.DocumentDTO
	ToModel(collection string, item map[string]interface{}) models.Document
	ToModels(collection string, items []map[string]interface{}) []models.Document
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
func (m *documentMapper) ToDTO(collection string, item map[string]interface{}) dto.DocumentDTO {
	dto := dto.DocumentDTO{}

	// Extract _id from document root
	if id, ok := item["_id"].(string); ok {
		dto.ID = id
	}

	// Extract nome and slug from root level
	if nome, ok := item["nome"].(string); ok {
		dto.Nome = nome
	}
	if slug, ok := item["slug"].(string); ok {
		dto.Slug = slug
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
func (m *documentMapper) ToDTOs(collection string, items []map[string]interface{}) []dto.DocumentDTO {
	documents := make([]dto.DocumentDTO, 0, len(items))

	for _, item := range items {
		doc := m.ToDTO(collection, item)
		documents = append(documents, doc)
	}

	return documents
}

// ToModel converts a single raw document to the existing models.Document format
func (m *documentMapper) ToModel(collection string, item map[string]interface{}) models.Document {
	model := models.Document{}

	// Extract _id from document root
	if id, ok := item["_id"].(string); ok {
		model.ID = id
	}

	// Extract nome and slug from root level
	if nome, ok := item["nome"].(string); ok {
		model.Nome = nome
	}
	if slug, ok := item["slug"].(string); ok {
		model.Slug = slug
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
func (m *documentMapper) ToModels(collection string, items []map[string]interface{}) []models.Document {
	documents := make([]models.Document, 0, len(items))

	for _, item := range items {
		doc := m.ToModel(collection, item)
		documents = append(documents, doc)
	}

	return documents
}