package product

import "errors"

var (
	ErrNameRequired    = errors.New("product name is required")
	ErrInvalidID       = errors.New("invalid product ID")
	ErrInvalidPrice    = errors.New("price must be greater than zero")
	ErrInvalidStock    = errors.New("stock cannot be negative")
	ErrProductNotFound = errors.New("product not found")
)
