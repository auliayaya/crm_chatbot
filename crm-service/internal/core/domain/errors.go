package domain

import "errors"

// Common errors
var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("resource conflict")
	ErrInternal     = errors.New("internal error")
)
