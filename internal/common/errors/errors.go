package errors

import "errors"

var (
	ErrConflict   = errors.New("conflicting resource already exist")
	ErrNotFound   = errors.New("resource doesn't exist")
	ErrNotAllowed = errors.New("this operation is not allowed")
)
