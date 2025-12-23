package product

import "time"

// CreateProductInput represents the input for creating a product
type CreateProductInput struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"gte=0"`
}

// Validate validates the create product input
func (i *CreateProductInput) Validate() error {
	if i.Name == "" {
		return ErrNameRequired
	}
	if i.Price <= 0 {
		return ErrInvalidPrice
	}
	if i.Stock < 0 {
		return ErrInvalidStock
	}
	return nil
}

// UpdateProductInput represents the input for updating a product
type UpdateProductInput struct {
	ID          string  `json:"id" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

// Validate validates the update product input
func (i *UpdateProductInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	if i.Name == "" {
		return ErrNameRequired
	}
	if i.Price <= 0 {
		return ErrInvalidPrice
	}
	return nil
}

// GetProductInput represents the input for getting a product
type GetProductInput struct {
	ID string `json:"id" validate:"required"`
}

// Validate validates the get product input
func (i *GetProductInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	return nil
}

// DeleteProductInput represents the input for deleting a product
type DeleteProductInput struct {
	ID string `json:"id" validate:"required"`
}

// Validate validates the delete product input
func (i *DeleteProductInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	return nil
}

// UpdateStockInput represents the input for updating product stock
type UpdateStockInput struct {
	ID       string `json:"id" validate:"required"`
	Quantity int    `json:"quantity" validate:"required"`
}

// Validate validates the update stock input
func (i *UpdateStockInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	return nil
}

// ListProductsInput represents the input for listing products
type ListProductsInput struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// Validate validates the list products input
func (i *ListProductsInput) Validate() error {
	if i.Limit <= 0 {
		i.Limit = 10
	}
	if i.Offset < 0 {
		i.Offset = 0
	}
	return nil
}

// ProductOutput represents the output for a product
type ProductOutput struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListProductsOutput represents the output for listing products
type ListProductsOutput struct {
	Products []*ProductOutput `json:"products"`
	Total    int64            `json:"total"`
}
