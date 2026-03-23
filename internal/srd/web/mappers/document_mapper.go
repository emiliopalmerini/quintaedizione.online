package mappers

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/display"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/dto"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
)

type DocumentMapper interface {
	ToDTO(collection string, doc *domain.Document) dto.DocumentDTO
	ToDTOs(collection string, docs []*domain.Document) []dto.DocumentDTO
	ToModel(collection string, doc *domain.Document) models.Document
	ToModels(collection string, docs []*domain.Document) []models.Document
}

type documentMapper struct {
	displayFactory *display.DisplayElementFactory
}

func NewDocumentMapper(displayFactory *display.DisplayElementFactory) DocumentMapper {
	return &documentMapper{
		displayFactory: displayFactory,
	}
}

func (m *documentMapper) ToDTO(collection string, doc *domain.Document) dto.DocumentDTO {
	d := dto.DocumentDTO{
		ID:    string(doc.ID),
		Title: doc.Title,
	}

	if translated, ok := doc.Fields["translated"].(bool); ok {
		d.Translated = translated
	}

	d.DisplayElements = m.displayFactory.GetDisplayElements(collection, doc)
	return d
}

func (m *documentMapper) ToDTOs(collection string, docs []*domain.Document) []dto.DocumentDTO {
	result := make([]dto.DocumentDTO, 0, len(docs))
	for _, doc := range docs {
		result = append(result, m.ToDTO(collection, doc))
	}
	return result
}

func (m *documentMapper) ToModel(collection string, doc *domain.Document) models.Document {
	model := models.Document{
		Title: doc.Title,
	}

	// Build composite ID (source/slug) for URL routing
	if doc.Source != "" {
		model.ID = doc.Source + "/" + string(doc.ID)
	} else {
		model.ID = string(doc.ID)
	}

	if translated, ok := doc.Fields["translated"].(bool); ok {
		model.Translated = translated
	}

	displayElements := m.displayFactory.GetDisplayElements(collection, doc)
	for _, elem := range displayElements {
		model.DisplayElements = append(model.DisplayElements, models.DocumentDisplayField{
			Value: elem.Value,
			Type:  elem.Type,
		})
	}

	return model
}

func (m *documentMapper) ToModels(collection string, docs []*domain.Document) []models.Document {
	result := make([]models.Document, 0, len(docs))
	for _, doc := range docs {
		result = append(result, m.ToModel(collection, doc))
	}
	return result
}
