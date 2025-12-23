package model

import (
	"time"
)

// Product domain errors are defined in domain_error.go

// Product represents a product in the catalog
type Product struct {
	ID          string // MongoDB ObjectID
	Name        string
	Description string
	Price       float64
	Stock       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	events []DomainEvent
}

// NewProduct creates a new product with validation
func NewProduct(name, description string, price float64, stock int) (*Product, error) {
	product := &Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := product.Validate(); err != nil {
		return nil, err
	}

	product.recordEvent(ProductCreatedEvent{
		Name:  name,
		Price: price,
		Stock: stock,
	})

	return product, nil
}

// Validate validates the product entity
func (p *Product) Validate() error {
	if p.Name == "" {
		return ErrProductNameRequired
	}

	if p.Price <= 0 {
		return ErrProductPriceInvalid
	}

	if p.Stock < 0 {
		return ErrProductStockNegative
	}

	return nil
}

// Update updates product information
func (p *Product) Update(name, description string, price float64) error {
	if name == "" {
		return ErrProductNameRequired
	}

	if price <= 0 {
		return ErrProductPriceInvalid
	}

	p.Name = name
	p.Description = description
	p.Price = price
	p.UpdatedAt = time.Now()

	p.recordEvent(ProductUpdatedEvent{
		ID:    p.ID,
		Name:  name,
		Price: price,
	})

	return nil
}

// UpdateStock updates the product stock
func (p *Product) UpdateStock(quantity int) error {
	newStock := p.Stock + quantity
	if newStock < 0 {
		return ErrProductStockNegative
	}

	p.Stock = newStock
	p.UpdatedAt = time.Now()

	p.recordEvent(StockUpdatedEvent{
		ProductID: p.ID,
		OldStock:  p.Stock - quantity,
		NewStock:  newStock,
		Change:    quantity,
	})

	return nil
}

// ReserveStock reserves stock for an order
func (p *Product) ReserveStock(quantity int) error {
	if p.Stock < quantity {
		return ErrProductInsufficientStock
	}

	p.Stock -= quantity
	p.UpdatedAt = time.Now()

	return nil
}

// MarkDeleted marks the product as deleted
func (p *Product) MarkDeleted() {
	now := time.Now()
	p.DeletedAt = &now

	p.recordEvent(ProductDeletedEvent{
		ID: p.ID,
	})
}

// Events returns and clears domain events
func (p *Product) Events() []DomainEvent {
	events := p.events
	p.events = nil
	return events
}

func (p *Product) recordEvent(event DomainEvent) {
	p.events = append(p.events, event)
}

// Product domain events
type ProductCreatedEvent struct {
	Name  string
	Price float64
	Stock int
}

func (e ProductCreatedEvent) EventName() string { return "product.created" }

type ProductUpdatedEvent struct {
	ID    string
	Name  string
	Price float64
}

func (e ProductUpdatedEvent) EventName() string { return "product.updated" }

type ProductDeletedEvent struct {
	ID string
}

func (e ProductDeletedEvent) EventName() string { return "product.deleted" }

type StockUpdatedEvent struct {
	ProductID string
	OldStock  int
	NewStock  int
	Change    int
}

func (e StockUpdatedEvent) EventName() string { return "product.stock_updated" }
