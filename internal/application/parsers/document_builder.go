package parsers

import (
	"strings"

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
