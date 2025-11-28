package domain

import (
	"errors"
	"fmt"
)

// Domain errors
var (
	ErrDocumentNotFound      = errors.New("document not found")
	ErrInvalidDocumentID     = errors.New("invalid document ID")
	ErrInvalidDocumentTitle  = errors.New("invalid document title")
	ErrDocumentAlreadyExists = errors.New("document already exists")
)

// DocumentError represents a domain-level error with additional context
type DocumentError struct {
	Op  string     // Operation that failed
	ID  DocumentID // Document ID if applicable
	Err error      // Underlying error
	Msg string     // Additional message
}

func (e *DocumentError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("document %s: %s %s: %v", e.ID, e.Op, e.Msg, e.Err)
	}
	return fmt.Sprintf("document: %s %s: %v", e.Op, e.Msg, e.Err)
}

func (e *DocumentError) Unwrap() error {
	return e.Err
}

// NewDocumentError creates a new DocumentError
func NewDocumentError(op string, id DocumentID, err error, msg string) *DocumentError {
	return &DocumentError{
		Op:  op,
		ID:  id,
		Err: err,
		Msg: msg,
	}
}

// IsDocumentNotFound checks if an error indicates a document was not found
func IsDocumentNotFound(err error) bool {
	if err == nil {
		return false
	}
	var docErr *DocumentError
	if errors.As(err, &docErr) {
		return errors.Is(docErr.Err, ErrDocumentNotFound)
	}
	return errors.Is(err, ErrDocumentNotFound)
}

// IsInvalidDocumentID checks if an error indicates an invalid document ID
func IsInvalidDocumentID(err error) bool {
	if err == nil {
		return false
	}
	var docErr *DocumentError
	if errors.As(err, &docErr) {
		return errors.Is(docErr.Err, ErrInvalidDocumentID)
	}
	return errors.Is(err, ErrInvalidDocumentID)
}
