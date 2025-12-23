package dto

import "time"

// CreateOrderReq represents the request to create an order
type CreateOrderReq struct {
	UserID string         `json:"user_id" binding:"required,uuid"`
	Items  []OrderItemReq `json:"items" binding:"required,min=1,dive"`
}

// OrderItemReq represents an order item in the request
type OrderItemReq struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gt=0"`
}

// UpdateOrderStatusReq represents the request to update order status
type UpdateOrderStatusReq struct {
	Status string `json:"status" binding:"required,oneof=confirmed shipped delivered canceled"`
}

// GetOrderReq represents the request to get an order
type GetOrderReq struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// CancelOrderReq represents the request to cancel an order
type CancelOrderReq struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// OrderResp represents the order response
type OrderResp struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Items     []OrderItemResp `json:"items"`
	Total     float64         `json:"total"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// OrderItemResp represents an order item in the response
type OrderItemResp struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
