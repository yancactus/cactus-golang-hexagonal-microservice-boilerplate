package order

import "time"

// OrderItemInput represents an order item in the input
type OrderItemInput struct {
	ProductID string  `json:"product_id" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	Price     float64 `json:"price" validate:"required,gt=0"`
}

// CreateOrderInput represents the input for creating an order
type CreateOrderInput struct {
	UserID string           `json:"user_id" validate:"required,uuid"`
	Items  []OrderItemInput `json:"items" validate:"required,min=1"`
}

// Validate validates the create order input
func (i *CreateOrderInput) Validate() error {
	if i.UserID == "" {
		return ErrInvalidUserID
	}
	if len(i.Items) == 0 {
		return ErrItemsRequired
	}
	for _, item := range i.Items {
		if item.ProductID == "" {
			return ErrInvalidProductID
		}
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.Price <= 0 {
			return ErrInvalidPrice
		}
	}
	return nil
}

// GetOrderInput represents the input for getting an order
type GetOrderInput struct {
	ID string `json:"id" validate:"required,uuid"`
}

// Validate validates the get order input
func (i *GetOrderInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	return nil
}

// UpdateStatusInput represents the input for updating order status
type UpdateStatusInput struct {
	ID     string `json:"id" validate:"required,uuid"`
	Status string `json:"status" validate:"required"`
}

// Validate validates the update status input
func (i *UpdateStatusInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	if i.Status == "" {
		return ErrStatusRequired
	}
	validStatuses := map[string]bool{
		"confirmed": true,
		"shipped":   true,
		"delivered": true,
		"canceled":  true,
	}
	if !validStatuses[i.Status] {
		return ErrInvalidStatus
	}
	return nil
}

// CancelOrderInput represents the input for canceling an order
type CancelOrderInput struct {
	ID string `json:"id" validate:"required,uuid"`
}

// Validate validates the cancel order input
func (i *CancelOrderInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	return nil
}

// ListOrdersInput represents the input for listing orders
type ListOrdersInput struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// Validate validates the list orders input
func (i *ListOrdersInput) Validate() error {
	if i.Limit <= 0 {
		i.Limit = 10
	}
	if i.Offset < 0 {
		i.Offset = 0
	}
	return nil
}

// GetUserOrdersInput represents the input for getting orders by user
type GetUserOrdersInput struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// Validate validates the get user orders input
func (i *GetUserOrdersInput) Validate() error {
	if i.UserID == "" {
		return ErrInvalidUserID
	}
	if i.Limit <= 0 {
		i.Limit = 10
	}
	if i.Offset < 0 {
		i.Offset = 0
	}
	return nil
}

// OrderItemOutput represents an order item in the output
type OrderItemOutput struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// OrderOutput represents the output for an order
type OrderOutput struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Items     []OrderItemOutput `json:"items"`
	Total     float64           `json:"total"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// ListOrdersOutput represents the output for listing orders
type ListOrdersOutput struct {
	Orders []*OrderOutput `json:"orders"`
	Total  int64          `json:"total"`
}
