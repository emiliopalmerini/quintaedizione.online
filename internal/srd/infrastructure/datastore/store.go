package datastore

import (
	"fmt"
	"sort"
	"strings"
)

// Store is a read-only in-memory document store.
// Documents are stored as map[string]any with at least "_id" and "title" fields.
type Store struct {
	collections map[string]map[string]map[string]any // collection → id → doc
	sorted      map[string][]string                  // collection → title-sorted IDs
	slugIndex   map[string]map[string][]string       // collection → bare slug → []compositeIDs
}

// NewStore builds a store from a map of collection name → documents.
func NewStore(data map[string][]map[string]any) *Store {
	s := &Store{
		collections: make(map[string]map[string]map[string]any, len(data)),
		sorted:      make(map[string][]string, len(data)),
		slugIndex:   make(map[string]map[string][]string, len(data)),
	}

	for collection, docs := range data {
		idMap := make(map[string]map[string]any, len(docs))
		for _, doc := range docs {
			id, _ := doc["_id"].(string)
			if id == "" {
				continue
			}
			// Use short_name/id as composite key for clean URLs
			short, _ := doc["_source_short"].(string)
			if short != "" {
				idMap[short+"/"+id] = doc
			} else {
				idMap[id] = doc
			}
		}
		s.collections[collection] = idMap

		// Build slug index: bare slug → []compositeIDs
		slugIdx := make(map[string][]string)
		for compositeID := range idMap {
			slug := compositeID
			if _, after, found := strings.Cut(compositeID, "/"); found {
				slug = after
			}
			slugIdx[slug] = append(slugIdx[slug], compositeID)
		}
		// Sort composite IDs within each slug for deterministic order
		for slug := range slugIdx {
			sort.Strings(slugIdx[slug])
		}
		s.slugIndex[collection] = slugIdx

		// Build title-sorted ID list
		type entry struct {
			id    string
			title string
		}
		entries := make([]entry, 0, len(idMap))
		for id, doc := range idMap {
			title, _ := doc["title"].(string)
			entries = append(entries, entry{id: id, title: strings.ToLower(title)})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].title < entries[j].title
		})
		ids := make([]string, len(entries))
		for i, e := range entries {
			ids[i] = e.id
		}
		s.sorted[collection] = ids
	}

	return s
}

// Get returns a document by collection and ID. Returns an error if not found.
func (s *Store) Get(collection, id string) (map[string]any, error) {
	coll, ok := s.collections[collection]
	if !ok {
		return nil, fmt.Errorf("collection %q not found", collection)
	}
	doc, ok := coll[id]
	if !ok {
		return nil, fmt.Errorf("document %q not found in collection %q", id, collection)
	}
	return doc, nil
}

// GetBySlug returns all documents in a collection that share the given bare slug
// across sources. Returns nil if the slug is not found.
func (s *Store) GetBySlug(collection, slug string) []map[string]any {
	slugIdx, ok := s.slugIndex[collection]
	if !ok {
		return nil
	}
	compositeIDs, ok := slugIdx[slug]
	if !ok {
		return nil
	}
	coll := s.collections[collection]
	docs := make([]map[string]any, 0, len(compositeIDs))
	for _, id := range compositeIDs {
		if doc, ok := coll[id]; ok {
			docs = append(docs, doc)
		}
	}
	return docs
}

// Query iterates over a collection, applies the match predicate, and returns
// a paginated slice plus the total count of matching documents.
// If match is nil, all documents match.
// Results are returned in title-sorted order.
func (s *Store) Query(collection string, match func(map[string]any) bool, skip, limit int64) ([]map[string]any, int64) {
	ids, ok := s.sorted[collection]
	if !ok {
		return nil, 0
	}
	coll := s.collections[collection]

	var matched []map[string]any
	for _, id := range ids {
		doc := coll[id]
		if match == nil || match(doc) {
			matched = append(matched, doc)
		}
	}

	total := int64(len(matched))

	// Apply pagination
	start := int(skip)
	if start >= len(matched) {
		return nil, total
	}
	matched = matched[start:]

	if limit > 0 && int64(len(matched)) > limit {
		matched = matched[:limit]
	}

	return matched, total
}

// Adjacent returns the previous and next document IDs relative to currentID
// in the title-sorted order of the collection, along with the 1-based position
// and total count of documents in the collection.
func (s *Store) Adjacent(collection, currentID string) (prevID, nextID *string, position, total int) {
	ids, ok := s.sorted[collection]
	if !ok {
		return nil, nil, 0, 0
	}

	idx := sort.SearchStrings(ids, currentID)

	// SearchStrings may not find exact match; verify
	if idx >= len(ids) || ids[idx] != currentID {
		// Fallback: linear search (IDs sorted by title, not alphabetically by ID)
		idx = -1
		for i, id := range ids {
			if id == currentID {
				idx = i
				break
			}
		}
		if idx == -1 {
			return nil, nil, 0, 0
		}
	}

	total = len(ids)
	position = idx + 1 // 1-based

	if idx > 0 {
		prevID = &ids[idx-1]
	}
	if idx < len(ids)-1 {
		nextID = &ids[idx+1]
	}

	return prevID, nextID, position, total
}

// Aggregate iterates over a collection, applies an optional predicate, reads the
// field at fieldPath from each matching document, and returns a map of value → count.
func (s *Store) Aggregate(collection, fieldPath string, match func(map[string]any) bool) map[string]int64 {
	ids, ok := s.sorted[collection]
	if !ok {
		return nil
	}
	coll := s.collections[collection]
	counts := make(map[string]int64)

	for _, id := range ids {
		doc := coll[id]
		if match != nil && !match(doc) {
			continue
		}
		val := getNestedField(doc, fieldPath)
		if val == nil {
			continue
		}
		counts[fmt.Sprintf("%v", val)]++
	}
	return counts
}

// getNestedField retrieves a potentially nested field from a document.
func getNestedField(doc map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = doc
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = m[part]
	}
	return current
}

// Count returns the number of documents in a collection.
func (s *Store) Count(collection string) int64 {
	return int64(len(s.collections[collection]))
}

// Collections returns the names of all collections.
func (s *Store) Collections() []string {
	names := make([]string, 0, len(s.collections))
	for name := range s.collections {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
