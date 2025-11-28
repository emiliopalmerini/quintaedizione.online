package filters

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoFilterBuilder struct{}

func NewMongoFilterBuilder() *MongoFilterBuilder {
	return &MongoFilterBuilder{}
}

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

func (b *MongoFilterBuilder) buildSingleFilter(filterValue filters.FilterValue) (bson.M, error) {
	def := filterValue.Definition
	value := filterValue.Value

	if value == "" {
		return bson.M{}, nil
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

func (b *MongoFilterBuilder) buildRegexMatch(fieldPath, value string) (bson.M, error) {
	escapedValue := regexp.QuoteMeta(value)
	return bson.M{
		fieldPath: bson.M{
			"$regex":   escapedValue,
			"$options": "i",
		},
	}, nil
}

func (b *MongoFilterBuilder) buildRangeMatch(fieldPath, value string, dataType filters.FilterDataType) (bson.M, error) {
	if dataType != filters.NumberFilter {
		return nil, fmt.Errorf("range match only supported for number filters")
	}

	value = strings.TrimSpace(value)

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

	numValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number value: %s", value)
	}
	return bson.M{fieldPath: numValue}, nil
}

func (b *MongoFilterBuilder) buildInMatch(fieldPath, value string) (bson.M, error) {

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

		return bson.M{fieldPath: trimmedValues[0]}, nil
	}

	return bson.M{fieldPath: bson.M{"$in": trimmedValues}}, nil
}

func (b *MongoFilterBuilder) BuildSearchFilter(collection filters.CollectionType, searchTerm string) bson.M {
	if searchTerm == "" {
		return bson.M{}
	}

	return bson.M{
		"$text": bson.M{
			"$search": searchTerm,
		},
	}
}
