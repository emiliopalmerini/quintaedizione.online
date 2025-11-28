package filters

import "fmt"

type ValidationError struct {
	FilterName string
	Message    string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("filter validation error for '%s': %s", e.FilterName, e.Message)
}

func NewValidationError(filterName, message string) ValidationError {
	return ValidationError{
		FilterName: filterName,
		Message:    message,
	}
}

type UnsupportedFilterError struct {
	FilterName string
	Collection CollectionType
}

func (e UnsupportedFilterError) Error() string {
	return fmt.Sprintf("filter '%s' is not supported for collection '%s'", e.FilterName, e.Collection)
}

func NewUnsupportedFilterError(filterName string, collection CollectionType) UnsupportedFilterError {
	return UnsupportedFilterError{
		FilterName: filterName,
		Collection: collection,
	}
}
