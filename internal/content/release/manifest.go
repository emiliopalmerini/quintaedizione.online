package release

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	manifestPath         = "manifest.json"
	supportedFormat      = 1
	supportedSchemaMajor = 1
	jsonMediaType        = "application/json"
)

var (
	semanticVersionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	datasetVersionPattern  = regexp.MustCompile(`^[0-9]{4}\.(0[1-9]|1[0-2])\.[1-9][0-9]*$`)
)

type Manifest struct {
	FormatVersion  int        `json:"formatVersion"`
	SchemaVersion  string     `json:"schemaVersion"`
	DatasetVersion string     `json:"datasetVersion"`
	Resources      []Resource `json:"resources"`
}

type Resource struct {
	Path       string `json:"path"`
	RecordKind string `json:"recordKind"`
	MediaType  string `json:"mediaType"`
	SHA256     string `json:"sha256"`
}

type Release struct {
	manifest  Manifest
	files     fs.FS
	resources map[string]Resource
}

func Load(files fs.FS) (Release, error) {
	data, err := fs.ReadFile(files, manifestPath)
	if err != nil {
		return Release{}, fmt.Errorf("read release manifest: %w", err)
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return Release{}, fmt.Errorf("decode release manifest: %w", err)
	}

	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Release{}, fmt.Errorf("decode release manifest: %w", err)
	}
	if err := rejectTrailingJSON(decoder); err != nil {
		return Release{}, fmt.Errorf("decode release manifest: %w", err)
	}
	if err := manifest.validate(); err != nil {
		return Release{}, err
	}

	resources := make(map[string]Resource, len(manifest.Resources))
	for _, resource := range manifest.Resources {
		if err := verifyResource(files, resource); err != nil {
			return Release{}, err
		}
		resources[resource.Path] = resource
	}
	return Release{manifest: manifest, files: files, resources: resources}, nil
}

func (release Release) Manifest() Manifest {
	manifest := release.manifest
	manifest.Resources = append([]Resource(nil), release.manifest.Resources...)
	return manifest
}

func (release Release) ReadResource(resourcePath string) ([]byte, error) {
	resource, exists := release.resources[resourcePath]
	if !exists {
		return nil, fmt.Errorf("resource %q is not declared by the release", resourcePath)
	}
	data, err := fs.ReadFile(release.files, resource.Path)
	if err != nil {
		return nil, fmt.Errorf("read release resource %q: %w", resource.Path, err)
	}
	if err := verifyChecksum(data, resource); err != nil {
		return nil, err
	}
	if err := verifyPayload(data, resource); err != nil {
		return nil, err
	}
	return data, nil
}

func (release Release) ResourcesByKind(recordKind string) []Resource {
	resources := make([]Resource, 0)
	for _, resource := range release.manifest.Resources {
		if resource.RecordKind == recordKind {
			resources = append(resources, resource)
		}
	}
	return resources
}

func (manifest Manifest) validate() error {
	if manifest.FormatVersion != supportedFormat {
		return fmt.Errorf("unsupported release format version %d", manifest.FormatVersion)
	}

	match := semanticVersionPattern.FindStringSubmatch(manifest.SchemaVersion)
	if match == nil {
		return fmt.Errorf("invalid canonical schema version %q", manifest.SchemaVersion)
	}
	major, err := strconv.Atoi(match[1])
	if err != nil || major != supportedSchemaMajor {
		return fmt.Errorf("unsupported canonical schema version %q", manifest.SchemaVersion)
	}
	if !datasetVersionPattern.MatchString(manifest.DatasetVersion) {
		return fmt.Errorf("invalid dataset version %q", manifest.DatasetVersion)
	}
	if len(manifest.Resources) == 0 {
		return errors.New("release manifest has no resources")
	}

	seen := make(map[string]struct{}, len(manifest.Resources))
	previousPath := ""
	for index, resource := range manifest.Resources {
		if err := resource.validate(); err != nil {
			return fmt.Errorf("resource %d: %w", index, err)
		}
		if _, exists := seen[resource.Path]; exists {
			return fmt.Errorf("resource %d: duplicate path %q", index, resource.Path)
		}
		if previousPath != "" && resource.Path < previousPath {
			return fmt.Errorf("resource %d: resources are not ordered by path", index)
		}
		seen[resource.Path] = struct{}{}
		previousPath = resource.Path
	}
	return nil
}

func (resource Resource) validate() error {
	if resource.Path == "" || resource.Path == "." || !fs.ValidPath(resource.Path) || path.Clean(resource.Path) != resource.Path || strings.Contains(resource.Path, `\`) {
		return fmt.Errorf("unsafe path %q", resource.Path)
	}
	if resource.Path == manifestPath {
		return errors.New("manifest cannot declare itself as a resource")
	}
	if strings.TrimSpace(resource.RecordKind) == "" {
		return errors.New("record kind is required")
	}
	if resource.MediaType != jsonMediaType {
		return fmt.Errorf("unsupported media type %q", resource.MediaType)
	}
	if len(resource.SHA256) != sha256.Size*2 || resource.SHA256 != strings.ToLower(resource.SHA256) {
		return fmt.Errorf("invalid SHA-256 digest %q", resource.SHA256)
	}
	if _, err := hex.DecodeString(resource.SHA256); err != nil {
		return fmt.Errorf("invalid SHA-256 digest %q", resource.SHA256)
	}
	return nil
}

func verifyResource(files fs.FS, resource Resource) error {
	data, err := fs.ReadFile(files, resource.Path)
	if err != nil {
		return fmt.Errorf("read release resource %q: %w", resource.Path, err)
	}
	if err := verifyChecksum(data, resource); err != nil {
		return err
	}
	return verifyPayload(data, resource)
}

func verifyChecksum(data []byte, resource Resource) error {
	actual := sha256.Sum256(data)
	if hex.EncodeToString(actual[:]) != resource.SHA256 {
		return fmt.Errorf("release resource %q: SHA-256 checksum mismatch", resource.Path)
	}
	return nil
}

func verifyPayload(data []byte, resource Resource) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return fmt.Errorf("release resource %q: payload must be a JSON array", resource.Path)
	}
	var records []json.RawMessage
	if err := json.Unmarshal(trimmed, &records); err != nil {
		return fmt.Errorf("release resource %q: decode records: %w", resource.Path, err)
	}
	for index, record := range records {
		record = bytes.TrimSpace(record)
		if len(record) == 0 || record[0] != '{' {
			return fmt.Errorf("release resource %q record %d: record must be a JSON object", resource.Path, index)
		}
		if err := rejectDuplicateJSONKeys(record); err != nil {
			return fmt.Errorf("release resource %q record %d: %w", resource.Path, index, err)
		}
		if resource.RecordKind == "tombstone" {
			continue
		}
		var envelope struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(record, &envelope); err != nil {
			return fmt.Errorf("release resource %q record %d: decode kind: %w", resource.Path, index, err)
		}
		if envelope.Kind != resource.RecordKind {
			return fmt.Errorf("release resource %q record %d has kind %q, want %q", resource.Path, index, envelope.Kind, resource.RecordKind)
		}
	}
	return nil
}

func rejectTrailingJSON(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	return rejectTrailingJSON(decoder)
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delimiter {
	case '{':
		keys := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, exists := keys[key]; exists {
				return fmt.Errorf("duplicate object key %q", key)
			}
			keys[key] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim('}') {
			return errors.New("unterminated JSON object")
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim(']') {
			return errors.New("unterminated JSON array")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	return nil
}
