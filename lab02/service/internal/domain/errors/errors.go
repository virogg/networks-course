package errors

import "errors"

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidProductName = errors.New("invalid product name")
	ErrInvalidProductID   = errors.New("invalid product id")
	ErrProductExists      = errors.New("product already exists")
)
