package dto

import "time"

// CreateProductReq represents the request to create a product
type CreateProductReq struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"gte=0"`
}

// UpdateProductReq represents the request to update a product
type UpdateProductReq struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
}

// UpdateStockReq represents the request to update product stock
type UpdateStockReq struct {
	Quantity int `json:"quantity" binding:"required"`
}

// GetProductReq represents the request to get a product
type GetProductReq struct {
	ID string `uri:"id" binding:"required"`
}

// DeleteProductReq represents the request to delete a product
type DeleteProductReq struct {
	ID string `uri:"id" binding:"required"`
}

// ProductResp represents the product response
type ProductResp struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
