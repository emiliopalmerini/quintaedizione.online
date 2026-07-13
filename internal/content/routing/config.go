package routing

import (
	"fmt"
	"io/fs"

	"github.com/emiliopalmerini/quintaedizione.online/internal/content/catalog"
	"github.com/emiliopalmerini/quintaedizione.online/internal/content/jsonstrict"
)

const supportedConfigVersion = 1

type routeConfig struct {
	Version int         `json:"version"`
	Routes  []wireEntry `json:"routes"`
}

type wireEntry struct {
	EntityID string   `json:"entityId"`
	Path     string   `json:"path"`
	Aliases  []string `json:"aliases,omitempty"`
}

func LoadRegistry(files fs.FS, configPath string, entities *catalog.Catalog) (*Registry, error) {
	data, err := fs.ReadFile(files, configPath)
	if err != nil {
		return nil, fmt.Errorf("read route registry %q: %w", configPath, err)
	}

	var config routeConfig
	if err := jsonstrict.Decode(data, &config); err != nil {
		return nil, fmt.Errorf("decode route registry %q: %w", configPath, err)
	}
	if config.Version != supportedConfigVersion {
		return nil, fmt.Errorf("unsupported route registry version %d", config.Version)
	}
	if len(config.Routes) == 0 {
		return nil, fmt.Errorf("route registry %q has no routes", configPath)
	}

	entries := make([]Entry, len(config.Routes))
	for index, route := range config.Routes {
		entries[index] = Entry{
			EntityID: route.EntityID,
			Path:     route.Path,
			Aliases:  append([]string(nil), route.Aliases...),
		}
	}
	return New(entities, entries)
}
