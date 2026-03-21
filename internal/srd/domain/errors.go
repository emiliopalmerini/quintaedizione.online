package domain

import (
	"errors"
	"fmt"
)

var (
	ErrDocumentNotFound      = errors.New("document not found")
	ErrInvalidDocumentID     = errors.New("invalid document ID")
	ErrInvalidDocumentTitle  = errors.New("invalid document title")
	ErrDocumentAlreadyExists = errors.New("document already exists")
)

type DocumentError struct {
	Op  string
	ID  DocumentID
	Err error
	Msg string
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

func NewDocumentError(op string, id DocumentID, err error, msg string) *DocumentError {
	return &DocumentError{
		Op:  op,
		ID:  id,
		Err: err,
		Msg: msg,
	}
}

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
