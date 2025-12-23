package model

import "net/http"

// DomainError represents a domain-specific error with HTTP status code mapping
type DomainError struct {
	Code       string // Machine-readable error code
	Message    string // Human-readable error message
	HTTPStatus int    // HTTP status code to return
}

// Error implements the error interface
func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string, httpStatus int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Common validation error codes
const (
	CodeValidationError   = "VALIDATION_ERROR"
	CodeNotFound          = "NOT_FOUND"
	CodeConflict          = "CONFLICT"
	CodeInvalidState      = "INVALID_STATE"
	CodeInsufficientStock = "INSUFFICIENT_STOCK"
)

// User domain errors
var (
	ErrUserNotFound         = NewDomainError("USER_NOT_FOUND", "user not found", http.StatusNotFound)
	ErrUserEmailRequired    = NewDomainError(CodeValidationError, "user email is required", http.StatusBadRequest)
	ErrUserEmailInvalid     = NewDomainError(CodeValidationError, "user email is invalid", http.StatusBadRequest)
	ErrUserNameRequired     = NewDomainError(CodeValidationError, "user name is required", http.StatusBadRequest)
	ErrUserPasswordRequired = NewDomainError(CodeValidationError, "user password is required", http.StatusBadRequest)
	ErrUserEmailTaken       = NewDomainError("EMAIL_TAKEN", "user email is already taken", http.StatusConflict)
)

// Product domain errors
var (
	ErrProductNotFound          = NewDomainError("PRODUCT_NOT_FOUND", "product not found", http.StatusNotFound)
	ErrProductNameRequired      = NewDomainError(CodeValidationError, "product name is required", http.StatusBadRequest)
	ErrProductPriceInvalid      = NewDomainError(CodeValidationError, "product price must be greater than zero", http.StatusBadRequest)
	ErrProductStockInvalid      = NewDomainError(CodeValidationError, "product stock cannot be negative", http.StatusBadRequest)
	ErrProductStockNegative     = NewDomainError(CodeValidationError, "product stock cannot be negative", http.StatusBadRequest)
	ErrProductInsufficientStock = NewDomainError(CodeInsufficientStock, "insufficient product stock", http.StatusConflict)
)

// Order domain errors
var (
	ErrOrderNotFound         = NewDomainError("ORDER_NOT_FOUND", "order not found", http.StatusNotFound)
	ErrOrderUserRequired     = NewDomainError(CodeValidationError, "order user is required", http.StatusBadRequest)
	ErrOrderItemsRequired    = NewDomainError(CodeValidationError, "order must have at least one item", http.StatusBadRequest)
	ErrOrderInvalidStatus    = NewDomainError(CodeInvalidState, "invalid order status transition", http.StatusBadRequest)
	ErrOrderAlreadyCancelled = NewDomainError(CodeInvalidState, "order is already canceled", http.StatusConflict)
	ErrOrderCannotCancel     = NewDomainError(CodeInvalidState, "order cannot be canceled in current status", http.StatusConflict)
	ErrOrderCannotConfirm    = NewDomainError(CodeInvalidState, "order cannot be confirmed in current status", http.StatusConflict)
	ErrOrderCannotShip       = NewDomainError(CodeInvalidState, "order cannot be shipped in current status", http.StatusConflict)
	ErrOrderCannotDeliver    = NewDomainError(CodeInvalidState, "order cannot be delivered in current status", http.StatusConflict)
)

// Audit domain errors
var (
	ErrAuditNotFound = NewDomainError("AUDIT_NOT_FOUND", "audit log not found", http.StatusNotFound)
)
