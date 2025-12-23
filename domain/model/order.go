package model

import (
	"time"

	"github.com/google/uuid"
)

// Order domain errors are defined in domain_error.go

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "canceled"
)

// Order represents an order in the system
type Order struct {
	ID        string
	UserID    string
	Items     []OrderItem
	Total     float64
	Status    OrderStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	events []DomainEvent
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	Price     float64
	CreatedAt time.Time
}

// NewOrder creates a new order with validation
func NewOrder(userID string, items []OrderItem) (*Order, error) {
	orderID := uuid.New().String()
	order := &Order{
		ID:        orderID,
		UserID:    userID,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Assign IDs to items
	for i := range items {
		items[i].ID = uuid.New().String()
		items[i].OrderID = orderID
		items[i].CreatedAt = time.Now()
	}
	order.Items = items

	if err := order.Validate(); err != nil {
		return nil, err
	}

	// Calculate total
	order.calculateTotal()

	order.recordEvent(OrderCreatedEvent{
		UserID:     userID,
		ItemCount:  len(items),
		TotalValue: order.Total,
	})

	return order, nil
}

// Validate validates the order entity
func (o *Order) Validate() error {
	if o.UserID == "" {
		return ErrOrderUserRequired
	}

	if len(o.Items) == 0 {
		return ErrOrderItemsRequired
	}

	return nil
}

// calculateTotal calculates the total order value
func (o *Order) calculateTotal() {
	var total float64
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	o.Total = total
}

// Confirm confirms the order
func (o *Order) Confirm() error {
	if o.Status != OrderStatusPending {
		return ErrOrderInvalidStatus
	}

	o.Status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()

	o.recordEvent(OrderStatusChangedEvent{
		OrderID:   o.ID,
		OldStatus: string(OrderStatusPending),
		NewStatus: string(OrderStatusConfirmed),
	})

	return nil
}

// Ship marks the order as shipped
func (o *Order) Ship() error {
	if o.Status != OrderStatusConfirmed {
		return ErrOrderInvalidStatus
	}

	o.Status = OrderStatusShipped
	o.UpdatedAt = time.Now()

	o.recordEvent(OrderStatusChangedEvent{
		OrderID:   o.ID,
		OldStatus: string(OrderStatusConfirmed),
		NewStatus: string(OrderStatusShipped),
	})

	return nil
}

// Deliver marks the order as delivered
func (o *Order) Deliver() error {
	if o.Status != OrderStatusShipped {
		return ErrOrderInvalidStatus
	}

	o.Status = OrderStatusDelivered
	o.UpdatedAt = time.Now()

	o.recordEvent(OrderStatusChangedEvent{
		OrderID:   o.ID,
		OldStatus: string(OrderStatusShipped),
		NewStatus: string(OrderStatusDelivered),
	})

	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if o.Status == OrderStatusCancelled {
		return ErrOrderAlreadyCancelled
	}

	if o.Status == OrderStatusDelivered {
		return ErrOrderInvalidStatus
	}

	oldStatus := o.Status
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()

	o.recordEvent(OrderCancelledEvent{
		OrderID:   o.ID,
		OldStatus: string(oldStatus),
	})

	return nil
}

// Events returns and clears domain events
func (o *Order) Events() []DomainEvent {
	events := o.events
	o.events = nil
	return events
}

func (o *Order) recordEvent(event DomainEvent) {
	o.events = append(o.events, event)
}

// TableName returns the table name for GORM
func (o *Order) TableName() string {
	return "orders"
}

// TableName returns the table name for GORM
func (i *OrderItem) TableName() string {
	return "order_items"
}

// Order domain events
type OrderCreatedEvent struct {
	UserID     string
	ItemCount  int
	TotalValue float64
}

func (e OrderCreatedEvent) EventName() string { return "order.created" }

type OrderStatusChangedEvent struct {
	OrderID   string
	OldStatus string
	NewStatus string
}

func (e OrderStatusChangedEvent) EventName() string { return "order.status_changed" }

type OrderCancelledEvent struct {
	OrderID   string
	OldStatus string
}

func (e OrderCancelledEvent) EventName() string { return "order.cancelled" }
