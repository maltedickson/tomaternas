package apperrors

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
