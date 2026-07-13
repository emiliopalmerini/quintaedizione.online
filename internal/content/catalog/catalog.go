package catalog

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/content/release"
)

type Entity struct {
	ID        string
	ConceptID string
	Kind      string
	Edition   string
	Name      string
	Revision  uint
	Source    Source
}

type Source struct {
	Scope    string
	Document string
	Locator  string
	Section  string
}

type Catalog struct {
	byID      map[string]Entity
	byConcept map[string][]Entity
	retired   map[string]string
}

type wireEntity struct {
	ID        string     `json:"id"`
	ConceptID string     `json:"conceptId"`
	Kind      string     `json:"kind"`
	Edition   string     `json:"edition"`
	Name      string     `json:"name"`
	Revision  uint       `json:"revision"`
	Source    wireSource `json:"source"`
}

type wireSource struct {
	Scope    string `json:"scope"`
	Document string `json:"document"`
	Locator  string `json:"locator"`
	Section  string `json:"section"`
}

type wireTombstone struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

func Compile(source release.Release) (*Catalog, error) {
	metadata := source.Manifest()
	if metadata.DatasetVersion == "" {
		return nil, errors.New("canonical release is not initialized")
	}
	catalog := &Catalog{
		byID:      make(map[string]Entity),
		byConcept: make(map[string][]Entity),
		retired:   make(map[string]string),
	}
	conceptKinds := make(map[string]string)
	conceptScopes := make(map[string]string)

	for _, resource := range metadata.Resources {
		data, err := source.ReadResource(resource.Path)
		if err != nil {
			return nil, err
		}
		trimmed := bytes.TrimSpace(data)
		if len(trimmed) == 0 || trimmed[0] != '[' {
			return nil, fmt.Errorf("decode %s resource %q: payload must be a JSON array", resource.RecordKind, resource.Path)
		}
		if resource.RecordKind == "tombstone" {
			var records []wireTombstone
			if err := json.Unmarshal(trimmed, &records); err != nil {
				return nil, fmt.Errorf("decode tombstone resource %q: %w", resource.Path, err)
			}
			for index, record := range records {
				if !validCanonicalID(record.ID) {
					return nil, fmt.Errorf("resource %q record %d: invalid tombstone ID %q", resource.Path, index, record.ID)
				}
				if strings.TrimSpace(record.Reason) == "" {
					return nil, fmt.Errorf("resource %q record %d: tombstone reason is required", resource.Path, index)
				}
				if _, exists := catalog.retired[record.ID]; exists {
					return nil, fmt.Errorf("duplicate tombstone ID %q", record.ID)
				}
				if _, exists := catalog.byID[record.ID]; exists {
					return nil, fmt.Errorf("identity %q is both active and retired", record.ID)
				}
				catalog.retired[record.ID] = record.Reason
			}
			continue
		}

		var records []wireEntity
		if err := json.Unmarshal(trimmed, &records); err != nil {
			return nil, fmt.Errorf("decode %s resource %q: %w", resource.RecordKind, resource.Path, err)
		}
		for index, record := range records {
			entity, err := projectEntity(record, resource.RecordKind)
			if err != nil {
				return nil, fmt.Errorf("resource %q record %d: %w", resource.Path, index, err)
			}
			if _, exists := catalog.byID[entity.ID]; exists {
				return nil, fmt.Errorf("duplicate entity ID %q", entity.ID)
			}
			if _, exists := catalog.retired[entity.ID]; exists {
				return nil, fmt.Errorf("identity %q is both active and retired", entity.ID)
			}
			if kind, exists := conceptKinds[entity.ConceptID]; exists && kind != entity.Kind {
				return nil, fmt.Errorf("concept %q contains kinds %q and %q", entity.ConceptID, kind, entity.Kind)
			}
			conceptKinds[entity.ConceptID] = entity.Kind

			scopeKey := strings.Join([]string{entity.ConceptID, entity.Edition, entity.Source.Scope}, "\x00")
			if existing, exists := conceptScopes[scopeKey]; exists {
				return nil, fmt.Errorf("entities %q and %q represent one concept in the same edition and source scope", existing, entity.ID)
			}
			conceptScopes[scopeKey] = entity.ID

			catalog.byID[entity.ID] = entity
			catalog.byConcept[entity.ConceptID] = append(catalog.byConcept[entity.ConceptID], entity)
		}
	}

	for conceptID := range catalog.byConcept {
		sort.Slice(catalog.byConcept[conceptID], func(i, j int) bool {
			left := catalog.byConcept[conceptID][i]
			right := catalog.byConcept[conceptID][j]
			if left.Edition != right.Edition {
				return left.Edition < right.Edition
			}
			if left.Source.Scope != right.Source.Scope {
				return left.Source.Scope < right.Source.Scope
			}
			return left.ID < right.ID
		})
	}
	return catalog, nil
}

func (catalog *Catalog) Entity(id string) (Entity, bool) {
	entity, exists := catalog.byID[id]
	return entity, exists
}

func (catalog *Catalog) Versions(conceptID string) []Entity {
	return append([]Entity(nil), catalog.byConcept[conceptID]...)
}

func (catalog *Catalog) Retired(id string) bool {
	_, exists := catalog.retired[id]
	return exists
}

func projectEntity(record wireEntity, recordKind string) (Entity, error) {
	switch {
	case !validCanonicalID(record.ID):
		return Entity{}, fmt.Errorf("invalid canonical entity ID %q", record.ID)
	case !validCanonicalID(record.ConceptID):
		return Entity{}, fmt.Errorf("entity %q: invalid concept ID %q", record.ID, record.ConceptID)
	case record.Kind != recordKind:
		return Entity{}, fmt.Errorf("entity %q has kind %q in a %q resource", record.ID, record.Kind, recordKind)
	case record.Edition != "5e" && record.Edition != "5.5e":
		return Entity{}, fmt.Errorf("entity %q: unsupported edition %q", record.ID, record.Edition)
	case strings.TrimSpace(record.Name) == "":
		return Entity{}, fmt.Errorf("entity %q: name is required", record.ID)
	case record.Revision == 0:
		return Entity{}, fmt.Errorf("entity %q: revision must be positive", record.ID)
	case strings.TrimSpace(record.Source.Scope) == "":
		return Entity{}, fmt.Errorf("entity %q: source scope is required", record.ID)
	case strings.TrimSpace(record.Source.Document) == "":
		return Entity{}, fmt.Errorf("entity %q: source document is required", record.ID)
	}

	return Entity{
		ID:        record.ID,
		ConceptID: record.ConceptID,
		Kind:      record.Kind,
		Edition:   record.Edition,
		Name:      record.Name,
		Revision:  record.Revision,
		Source: Source{
			Scope:    record.Source.Scope,
			Document: record.Source.Document,
			Locator:  record.Source.Locator,
			Section:  record.Source.Section,
		},
	}, nil
}

func validCanonicalID(value string) bool {
	if len(value) != 36 || value != strings.ToLower(value) || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	compact := strings.ReplaceAll(value, "-", "")
	decoded := make([]byte, 16)
	if _, err := hex.Decode(decoded, []byte(compact)); err != nil {
		return false
	}
	return decoded[6]>>4 == 4 && decoded[8]&0xc0 == 0x80
}
