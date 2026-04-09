package services

import (
	"context"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/repositories"
)

// mockDocumentRepository implements all DocumentRepository interfaces for testing.
type mockDocumentRepository struct {
	findVersionsFn func(collection, slug string) ([]domain.VersionInfo, error)
}

func (m *mockDocumentRepository) FindByID(_ context.Context, _ string, _ string) (*domain.Document, error) {
	return nil, nil
}
func (m *mockDocumentRepository) FindByPredicate(_ context.Context, _ string, _ repositories.DocumentPredicate, _ int64, _ int64) ([]*domain.Document, int64, error) {
	return nil, 0, nil
}
func (m *mockDocumentRepository) Count(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (m *mockDocumentRepository) GetAllCollectionStats(_ context.Context) ([]map[string]any, error) {
	return nil, nil
}
func (m *mockDocumentRepository) AggregateField(_ context.Context, _, _ string, _ repositories.DocumentPredicate) (map[string]int64, error) {
	return nil, nil
}
func (m *mockDocumentRepository) GetAdjacent(_ context.Context, _ string, _ string) (*string, *string, int, int, error) {
	return nil, nil, 0, 0, nil
}
func (m *mockDocumentRepository) FindVersions(_ context.Context, collection, slug string) ([]domain.VersionInfo, error) {
	if m.findVersionsFn != nil {
		return m.findVersionsFn(collection, slug)
	}
	return nil, nil
}

// noopFilterService is a minimal FilterService for service tests.
type noopFilterService struct{}

func (n *noopFilterService) ParseFilters(_ collections.CollectionName, _ map[string]string) (*filters.FilterSet, error) {
	return &filters.FilterSet{}, nil
}
func (n *noopFilterService) ValidateFilterSet(_ *filters.FilterSet) error { return nil }
func (n *noopFilterService) BuildFilter(_ *filters.FilterSet) (filters.DocumentPredicate, error) {
	return nil, nil
}
func (n *noopFilterService) GetAvailableFilters(_ collections.CollectionName) ([]filters.FilterDefinition, error) {
	return nil, nil
}
func (n *noopFilterService) BuildSearchPredicate(_ collections.CollectionName, _ string) filters.DocumentPredicate {
	return nil
}
func (n *noopFilterService) CombinePredicates(_ ...filters.DocumentPredicate) filters.DocumentPredicate {
	return nil
}

func TestGetItemVersions_DelegatesToRepository(t *testing.T) {
	expectedVersions := []domain.VersionInfo{
		{SourceShort: "5.5e", CompositeID: "5.5e/palla-di-fuoco"},
		{SourceShort: "5e", CompositeID: "5e/palla-di-fuoco"},
	}
	repo := &mockDocumentRepository{
		findVersionsFn: func(collection, slug string) ([]domain.VersionInfo, error) {
			if collection != "incantesimi" || slug != "palla-di-fuoco" {
				t.Errorf("unexpected args: collection=%q, slug=%q", collection, slug)
			}
			return expectedVersions, nil
		},
	}
	svc := NewContentService(repo, &noopFilterService{})

	versions, err := svc.GetItemVersions(context.Background(), "incantesimi", "palla-di-fuoco")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].SourceShort != "5.5e" {
		t.Errorf("expected first version 5.5e, got %s", versions[0].SourceShort)
	}
}
