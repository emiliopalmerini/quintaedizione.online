package datastore

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// LoadJSON reads a JSON file from an fs.FS and decodes it into a slice of T.
func LoadJSON[T any](fsys fs.FS, path string) ([]T, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	return items, nil
}
