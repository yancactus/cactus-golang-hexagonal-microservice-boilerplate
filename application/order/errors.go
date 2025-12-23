package order

import "errors"

var (
	ErrInvalidID        = errors.New("invalid order ID")
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrInvalidProductID = errors.New("invalid product ID")
	ErrInvalidQuantity  = errors.New("quantity must be greater than zero")
	ErrInvalidPrice     = errors.New("price must be greater than zero")
	ErrItemsRequired    = errors.New("at least one item is required")
	ErrStatusRequired   = errors.New("status is required")
	ErrInvalidStatus    = errors.New("invalid status value")
	ErrOrderNotFound    = errors.New("order not found")
)
