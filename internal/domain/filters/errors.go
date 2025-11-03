package filters

import "fmt"

// ValidationError represents a filter validation error
type ValidationError struct {
	FilterName string
	Message    string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("filter validation error for '%s': %s", e.FilterName, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(filterName, message string) ValidationError {
	return ValidationError{
		FilterName: filterName,
		Message:    message,
	}
}

// UnsupportedFilterError represents an error for unsupported filters
type UnsupportedFilterError struct {
	FilterName string
	Collection CollectionType
}

func (e UnsupportedFilterError) Error() string {
	return fmt.Sprintf("filter '%s' is not supported for collection '%s'", e.FilterName, e.Collection)
}

// NewUnsupportedFilterError creates a new unsupported filter error
func NewUnsupportedFilterError(filterName string, collection CollectionType) UnsupportedFilterError {
	return UnsupportedFilterError{
		FilterName: filterName,
		Collection: collection,
	}
}
