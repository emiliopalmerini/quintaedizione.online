package mappers

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/display"
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/dto"
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/models"
)

type DocumentMapper interface {
	ToDTO(collection string, item map[string]any) dto.DocumentDTO
	ToDTOs(collection string, items []map[string]any) []dto.DocumentDTO
	ToModel(collection string, item map[string]any) models.Document
	ToModels(collection string, items []map[string]any) []models.Document
}

type documentMapper struct {
	displayFactory *display.DisplayElementFactory
}

func NewDocumentMapper(displayFactory *display.DisplayElementFactory) DocumentMapper {
	return &documentMapper{
		displayFactory: displayFactory,
	}
}

func (m *documentMapper) ToDTO(collection string, item map[string]any) dto.DocumentDTO {
	dto := dto.DocumentDTO{}

	if id, ok := item["_id"].(string); ok {
		dto.ID = id
	}

	if title, ok := item["title"].(string); ok {
		dto.Title = title
	}

	if translated, ok := item["translated"].(bool); ok {
		dto.Translated = translated
	}

	dto.DisplayElements = m.displayFactory.GetDisplayElements(collection, item)

	return dto
}

func (m *documentMapper) ToDTOs(collection string, items []map[string]any) []dto.DocumentDTO {
	documents := make([]dto.DocumentDTO, 0, len(items))

	for _, item := range items {
		doc := m.ToDTO(collection, item)
		documents = append(documents, doc)
	}

	return documents
}

func (m *documentMapper) ToModel(collection string, item map[string]any) models.Document {
	model := models.Document{}

	if id, ok := item["_id"].(string); ok {
		model.ID = id
	}

	if title, ok := item["title"].(string); ok {
		model.Title = title
	}

	if translated, ok := item["translated"].(bool); ok {
		model.Translated = translated
	}

	displayElements := m.displayFactory.GetDisplayElements(collection, item)
	for _, elem := range displayElements {
		model.DisplayElements = append(model.DisplayElements, models.DocumentDisplayField{
			Value: elem.Value,
		})
	}

	return model
}

func (m *documentMapper) ToModels(collection string, items []map[string]any) []models.Document {
	documents := make([]models.Document, 0, len(items))

	for _, item := range items {
		doc := m.ToModel(collection, item)
		documents = append(documents, doc)
	}

	return documents
}
