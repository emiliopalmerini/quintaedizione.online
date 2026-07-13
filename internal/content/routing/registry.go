package routing

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/content/catalog"
)

type Entry struct {
	EntityID string
	Path     string
	Aliases  []string
}

type Resolved struct {
	EntityID      string
	CanonicalPath string
	Redirect      bool
}

type Registry struct {
	byEntity map[string]Entry
	byPath   map[string]Resolved
}

func New(entities *catalog.Catalog, entries []Entry) (*Registry, error) {
	if entities == nil {
		return nil, errors.New("content catalog is required")
	}
	registry := &Registry{
		byEntity: make(map[string]Entry, len(entries)),
		byPath:   make(map[string]Resolved),
	}

	for index, entry := range entries {
		if _, exists := entities.Entity(entry.EntityID); !exists {
			return nil, fmt.Errorf("route %d references unknown entity %q", index, entry.EntityID)
		}
		if _, exists := registry.byEntity[entry.EntityID]; exists {
			return nil, fmt.Errorf("entity %q has multiple canonical routes", entry.EntityID)
		}
		if err := validatePath(entry.Path); err != nil {
			return nil, fmt.Errorf("route %d: canonical path: %w", index, err)
		}

		stored := Entry{
			EntityID: entry.EntityID,
			Path:     entry.Path,
			Aliases:  append([]string(nil), entry.Aliases...),
		}
		if err := registry.registerPath(stored.Path, Resolved{
			EntityID:      stored.EntityID,
			CanonicalPath: stored.Path,
		}); err != nil {
			return nil, fmt.Errorf("route %d: %w", index, err)
		}
		for aliasIndex, alias := range stored.Aliases {
			if err := validatePath(alias); err != nil {
				return nil, fmt.Errorf("route %d alias %d: %w", index, aliasIndex, err)
			}
			if err := registry.registerPath(alias, Resolved{
				EntityID:      stored.EntityID,
				CanonicalPath: stored.Path,
				Redirect:      true,
			}); err != nil {
				return nil, fmt.Errorf("route %d alias %d: %w", index, aliasIndex, err)
			}
		}
		registry.byEntity[stored.EntityID] = stored
	}
	return registry, nil
}

func (registry *Registry) Route(entityID string) (Entry, bool) {
	entry, exists := registry.byEntity[entityID]
	entry.Aliases = append([]string(nil), entry.Aliases...)
	return entry, exists
}

func (registry *Registry) Resolve(requestPath string) (Resolved, bool) {
	resolved, exists := registry.byPath[requestPath]
	return resolved, exists
}

func (registry *Registry) registerPath(routePath string, resolved Resolved) error {
	if existing, exists := registry.byPath[routePath]; exists {
		return fmt.Errorf("path %q collides with route for entity %q", routePath, existing.EntityID)
	}
	registry.byPath[routePath] = resolved
	return nil
}

func validatePath(routePath string) error {
	if routePath == "" || routePath == "/" || !strings.HasPrefix(routePath, "/") {
		return fmt.Errorf("invalid absolute path %q", routePath)
	}
	if strings.Contains(routePath, `\`) {
		return fmt.Errorf("path %q cannot contain backslashes", routePath)
	}
	if strings.ContainsAny(routePath, "?#") || path.Clean(routePath) != routePath {
		return fmt.Errorf("path %q must be canonical and cannot contain a query or fragment", routePath)
	}
	parsed, err := url.ParseRequestURI(routePath)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return fmt.Errorf("invalid route path %q", routePath)
	}
	if parsed.Path != routePath || path.Clean(parsed.Path) != parsed.Path {
		return fmt.Errorf("path %q must not contain encoded or normalized segments", routePath)
	}
	return nil
}
