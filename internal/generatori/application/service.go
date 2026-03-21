package application

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"sort"

	"github.com/emiliopalmerini/quintaedizione.online/internal/generatori/domain"
)

// Service provides random generator operations.
type Service struct {
	tables map[string]domain.Table
	order  []string // sorted IDs for stable listing
}

// NewService loads all generator tables from the embedded filesystem.
func NewService(dataFS fs.FS) (*Service, error) {
	tables := make(map[string]domain.Table)

	entries, err := fs.ReadDir(dataFS, ".")
	if err != nil {
		return nil, fmt.Errorf("reading generator data dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "embed.go" {
			continue
		}

		data, err := fs.ReadFile(dataFS, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var t domain.Table
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		if err := validateTable(t, entry.Name()); err != nil {
			return nil, err
		}

		tables[t.ID] = t
	}

	order := make([]string, 0, len(tables))
	for id := range tables {
		order = append(order, id)
	}
	sort.Slice(order, func(i, j int) bool {
		return tables[order[i]].Order < tables[order[j]].Order
	})

	return &Service{tables: tables, order: order}, nil
}

func validateTable(t domain.Table, filename string) error {
	if t.IsMultiColumn() {
		for _, col := range t.Columns {
			if len(col.Items) == 0 {
				return fmt.Errorf("table %s column %q has no items", filename, col.Name)
			}
		}
		return nil
	}
	if len(t.Items) == 0 {
		return fmt.Errorf("table %s has no items", filename)
	}
	return nil
}

// List returns all tables in stable order.
func (s *Service) List() []domain.Table {
	result := make([]domain.Table, 0, len(s.order))
	for _, id := range s.order {
		result = append(result, s.tables[id])
	}
	return result
}

// ListGroups returns tables organized into groups, sorted by group order then table order.
func (s *Service) ListGroups() []domain.Group {
	// Collect tables by group ID.
	byGroup := make(map[string][]domain.Table)
	for _, id := range s.order {
		t := s.tables[id]
		byGroup[t.Group] = append(byGroup[t.Group], t)
	}

	// Build groups using the registry order.
	var groups []domain.Group
	for _, gi := range domain.GroupRegistry {
		tables, ok := byGroup[gi.ID]
		if !ok {
			continue
		}
		groups = append(groups, domain.Group{
			ID:          gi.ID,
			Label:       gi.Label,
			Description: gi.Description,
			Tables:      tables,
		})
		delete(byGroup, gi.ID)
	}

	// Append any tables with unknown group (shouldn't happen, but safe fallback).
	for groupID, tables := range byGroup {
		groups = append(groups, domain.Group{
			ID:     groupID,
			Label:  groupID,
			Tables: tables,
		})
	}

	return groups
}

// Get returns a single table by ID.
func (s *Service) Get(id string) (domain.Table, bool) {
	t, ok := s.tables[id]
	return t, ok
}

// Neighbors returns the previous and next table relative to the given ID.
func (s *Service) Neighbors(id string) (prev, next *domain.Table) {
	for i, oid := range s.order {
		if oid == id {
			if i > 0 {
				t := s.tables[s.order[i-1]]
				prev = &t
			}
			if i < len(s.order)-1 {
				t := s.tables[s.order[i+1]]
				next = &t
			}
			return
		}
	}
	return
}

// Roll returns a random result from the given table.
func (s *Service) Roll(id string) (domain.RollResult, error) {
	t, ok := s.tables[id]
	if !ok {
		return domain.RollResult{}, fmt.Errorf("table %q not found", id)
	}

	if t.Static {
		return domain.RollResult{}, fmt.Errorf("table %q is static and cannot be rolled", id)
	}

	if t.IsMultiColumn() {
		entries := make([]domain.RollEntry, len(t.Columns))
		for i, col := range t.Columns {
			item := col.Items[rand.IntN(len(col.Items))]
			entries[i] = domain.RollEntry{
				Column: col.Name,
				Value:  item.Text,
				Link:   item.Link,
			}
		}
		return domain.RollResult{Entries: entries}, nil
	}

	item := t.Items[rand.IntN(len(t.Items))]
	return domain.RollResult{
		Entries: []domain.RollEntry{
			{Value: item.Text, Link: item.Link},
		},
	}, nil
}
