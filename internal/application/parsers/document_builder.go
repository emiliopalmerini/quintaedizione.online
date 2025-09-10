package parsers

import (
	"fmt"
	"strings"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type BaseDocumentData struct {
	Nome        string
	Slug        domain.Slug
	Descrizione string
	Lingua      string
	Fonte       string
	FileOrigine string
}

type DocumentBuilder struct {
	baseData *BaseDocumentData
}

func NewDocumentBuilder(title, content string, context *ParsingContext) (*DocumentBuilder, error) {
	slug, err := domain.NewSlug(title)
	if err != nil {
		return nil, err
	}

	return &DocumentBuilder{
		baseData: &BaseDocumentData{
			Nome:        title,
			Slug:        slug,
			Descrizione: strings.TrimSpace(content),
			Lingua:      context.Language,
			Fonte:       "SRD 5.2",
			FileOrigine: context.Filename,
		},
	}, nil
}

func (db *DocumentBuilder) WithMetadata(key, value string) *DocumentBuilder {
	// Allow chaining for additional fields
	return db
}

func (db *DocumentBuilder) BuildDocumento() *domain.Documento {
	return &domain.Documento{
		Slug:      db.baseData.Slug,
		Titolo:    db.baseData.Nome,
		Contenuto: db.baseData.Descrizione,
	}
}

func (db *DocumentBuilder) GetBaseData() *BaseDocumentData {
	return db.baseData
}

// BuildDocument creates a document suitable for persistence from a ParsedEntity
func (db *DocumentBuilder) BuildDocument(entity domain.ParsedEntity, collection string) (map[string]any, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity cannot be nil")
	}

	// Create base document structure
	doc := map[string]any{
		"entity_type": entity.EntityType(),
		"value":       entity,
		"collection":  collection,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}

	// Add metadata from base data if available
	if db.baseData != nil {
		doc["language"] = db.baseData.Lingua
		doc["source"] = db.baseData.Fonte
		doc["source_file"] = db.baseData.FileOrigine
	} else {
		// Default values
		doc["language"] = "ita"
		doc["source"] = "SRD 5.2"
	}

	return doc, nil
}
