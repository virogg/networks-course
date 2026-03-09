package errors

import "errors"

var (
	ErrValidation          = errors.New("validation")
	ErrInvalidInput        = errors.New("invalid input")
	ErrNoAvailableCouriers = errors.New("conflict")
	ErrAlreadyExists       = errors.New("resource already exists")
	ErrNotFound            = errors.New("resource not found")
)
