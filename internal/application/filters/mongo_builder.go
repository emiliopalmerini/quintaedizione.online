package filters

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/filters"
	"go.mongodb.org/mongo-driver/bson"
)

// MongoFilterBuilder builds MongoDB queries from filter sets
type MongoFilterBuilder struct{}

// NewMongoFilterBuilder creates a new MongoDB filter builder
func NewMongoFilterBuilder() *MongoFilterBuilder {
	return &MongoFilterBuilder{}
}

// BuildFilter builds a MongoDB filter from a filter set
func (b *MongoFilterBuilder) BuildFilter(filterSet *filters.FilterSet) (bson.M, error) {
	if !filterSet.HasFilters() {
		return bson.M{}, nil
	}

	var conditions []bson.M

	for _, filterValue := range filterSet.Filters {
		condition, err := b.buildSingleFilter(filterValue)
		if err != nil {
			return nil, fmt.Errorf("failed to build filter for %s: %w", filterValue.Definition.Name, err)
		}

		if len(condition) > 0 {
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return bson.M{}, nil
	} else if len(conditions) == 1 {
		return conditions[0], nil
	} else {
		return bson.M{"$and": conditions}, nil
	}
}

// buildSingleFilter builds a MongoDB condition for a single filter
func (b *MongoFilterBuilder) buildSingleFilter(filterValue filters.FilterValue) (bson.M, error) {
	def := filterValue.Definition
	value := filterValue.Value

	if value == "" {
		return bson.M{}, nil // Skip empty values
	}

	switch def.Operator {
	case filters.ExactMatch:
		return b.buildExactMatch(def.FieldPath, value, def.DataType)
	case filters.RegexMatch:
		return b.buildRegexMatch(def.FieldPath, value)
	case filters.RangeMatch:
		return b.buildRangeMatch(def.FieldPath, value, def.DataType)
	case filters.InMatch:
		return b.buildInMatch(def.FieldPath, value)
	default:
		return nil, fmt.Errorf("unsupported operator: %d", def.Operator)
	}
}

// buildExactMatch builds an exact match condition
func (b *MongoFilterBuilder) buildExactMatch(fieldPath, value string, dataType filters.FilterDataType) (bson.M, error) {
	switch dataType {
	case filters.StringFilter, filters.EnumFilter:
		return bson.M{fieldPath: value}, nil
	case filters.NumberFilter:
		numValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number value: %s", value)
		}
		return bson.M{fieldPath: numValue}, nil
	case filters.BooleanFilter:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid boolean value: %s", value)
		}
		return bson.M{fieldPath: boolValue}, nil
	default:
		return nil, fmt.Errorf("unsupported data type for exact match: %d", dataType)
	}
}

// buildRegexMatch builds a case-insensitive regex match condition
func (b *MongoFilterBuilder) buildRegexMatch(fieldPath, value string) (bson.M, error) {
	escapedValue := regexp.QuoteMeta(value)
	return bson.M{
		fieldPath: bson.M{
			"$regex":   escapedValue,
			"$options": "i",
		},
	}, nil
}

// buildRangeMatch builds a range match condition (for numeric values)
func (b *MongoFilterBuilder) buildRangeMatch(fieldPath, value string, dataType filters.FilterDataType) (bson.M, error) {
	if dataType != filters.NumberFilter {
		return nil, fmt.Errorf("range match only supported for number filters")
	}

	// Parse range formats like "100-500", ">100", "<500", ">=100", "<=500"
	value = strings.TrimSpace(value)

	// Handle comparison operators
	if strings.HasPrefix(value, ">=") {
		numValue, err := strconv.ParseFloat(strings.TrimSpace(value[2:]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range value: %s", value)
		}
		return bson.M{fieldPath: bson.M{"$gte": numValue}}, nil
	}

	if strings.HasPrefix(value, "<=") {
		numValue, err := strconv.ParseFloat(strings.TrimSpace(value[2:]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range value: %s", value)
		}
		return bson.M{fieldPath: bson.M{"$lte": numValue}}, nil
	}

	if strings.HasPrefix(value, ">") {
		numValue, err := strconv.ParseFloat(strings.TrimSpace(value[1:]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range value: %s", value)
		}
		return bson.M{fieldPath: bson.M{"$gt": numValue}}, nil
	}

	if strings.HasPrefix(value, "<") {
		numValue, err := strconv.ParseFloat(strings.TrimSpace(value[1:]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range value: %s", value)
		}
		return bson.M{fieldPath: bson.M{"$lt": numValue}}, nil
	}

	// Handle range format like "100-500"
	if strings.Contains(value, "-") {
		parts := strings.Split(value, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format: %s", value)
		}

		minValue, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range min value: %s", parts[0])
		}

		maxValue, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range max value: %s", parts[1])
		}

		return bson.M{
			fieldPath: bson.M{
				"$gte": minValue,
				"$lte": maxValue,
			},
		}, nil
	}

	// Handle exact number (fallback to exact match)
	numValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number value: %s", value)
	}
	return bson.M{fieldPath: numValue}, nil
}

// buildInMatch builds an IN condition for multiple values
func (b *MongoFilterBuilder) buildInMatch(fieldPath, value string) (bson.M, error) {
	// Split by comma and trim whitespace
	values := strings.Split(value, ",")
	trimmedValues := make([]string, 0, len(values))

	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			trimmedValues = append(trimmedValues, trimmed)
		}
	}

	if len(trimmedValues) == 0 {
		return bson.M{}, nil
	}

	if len(trimmedValues) == 1 {
		// Single value, use exact match
		return bson.M{fieldPath: trimmedValues[0]}, nil
	}

	// Multiple values, use $in
	return bson.M{fieldPath: bson.M{"$in": trimmedValues}}, nil
}

// BuildSearchFilter builds a fuzzy text search filter (separate from field filters)
func (b *MongoFilterBuilder) BuildSearchFilter(collection filters.CollectionType, searchTerm string) bson.M {
	if searchTerm == "" {
		return bson.M{}
	}

	// Build fuzzy regex pattern - allows for partial matches and word boundaries
	fuzzyPattern := b.buildFuzzyPattern(searchTerm)

	// Base search fields that apply to all collections
	baseSearchFields := []bson.M{
		{"title": bson.M{"$regex": fuzzyPattern, "$options": "i"}},
		{"content": bson.M{"$regex": fuzzyPattern, "$options": "i"}},
	}

	// Collection-specific search fields
	collectionSpecificFields := b.getCollectionSpecificSearchFields(collection, fuzzyPattern)

	// Combine base and collection-specific fields
	searchFields := append(baseSearchFields, collectionSpecificFields...)

	return bson.M{"$or": searchFields}
}

// buildFuzzyPattern builds a fuzzy search regex pattern
// Splits the search term into words and allows partial matching
func (b *MongoFilterBuilder) buildFuzzyPattern(searchTerm string) string {
	// Trim and normalize whitespace
	searchTerm = strings.TrimSpace(searchTerm)
	if searchTerm == "" {
		return ""
	}

	// Split into words
	words := strings.Fields(searchTerm)
	if len(words) == 0 {
		return ""
	}

	// Build pattern for each word
	var patterns []string
	for _, word := range words {
		// Escape special regex characters but allow for fuzzy matching
		escaped := regexp.QuoteMeta(word)
		// Allow word to appear anywhere (not just at word boundaries)
		patterns = append(patterns, escaped)
	}

	// Join all word patterns with .* to allow any characters between them
	// This creates a pattern like: (?=.*word1)(?=.*word2) for AND matching
	// Or just join them for sequential fuzzy matching
	if len(patterns) == 1 {
		return patterns[0]
	}

	// Use sequential matching for multi-word queries (more intuitive)
	return strings.Join(patterns, ".*")
}

// getCollectionSpecificSearchFields returns search fields specific to each collection
func (b *MongoFilterBuilder) getCollectionSpecificSearchFields(collection filters.CollectionType, escapedSearch string) []bson.M {
	switch collection {
	case filters.IncantesimiCollection:
		return []bson.M{
			{"scuola": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"livello": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"classi": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	case filters.MostriCollection, filters.AnimaliCollection:
		return []bson.M{
			{"tipo": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"taglia": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"gs": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"cr": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"grado_sfida": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"ambiente": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"allineamento": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	case filters.ArmiCollection:
		return []bson.M{
			{"categoria": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"tipo_danno": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"proprieta": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	case filters.ArmatureCollection:
		return []bson.M{
			{"categoria": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"tipo": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	case filters.OggettiMagiciCollection:
		return []bson.M{
			{"tipo": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"rarita": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"sintonia": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	case filters.ClassiCollection:
		return []bson.M{
			{"dado_vita": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"abilita_primaria": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"tiri_salvezza": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	case filters.BackgroundsCollection:
		return []bson.M{
			{"competenze_abilita": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"competenze_linguaggi": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"competenze_strumenti": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	case filters.TalentiCollection:
		return []bson.M{
			{"categoria": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"prerequisiti": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	default:
		return []bson.M{
			{"descrizione": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"categoria": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"tipo": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	}
}
