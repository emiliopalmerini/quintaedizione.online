package filters

import (
	"fmt"
	"strconv"
)

// ConvertValue converts a string value to the appropriate type based on the FilterDataType.
// Returns the converted value or an error if conversion fails.
func ConvertValue(value string, dataType FilterDataType) (any, error) {
	switch dataType {
	case StringFilter, EnumFilter:
		return value, nil
	case NumberFilter:
		// Try parsing as float first (more general)
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue, nil
		}
		// Try parsing as int
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue, nil
		}
		return nil, fmt.Errorf("invalid number format: %s", value)
	case BooleanFilter:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid boolean format: %s", value)
		}
		return boolValue, nil
	default:
		return value, nil
	}
}
