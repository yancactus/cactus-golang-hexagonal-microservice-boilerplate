package postgre

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
)

// OrderRepository implements IOrderRepo using PostgreSQL
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *gorm.DB) repo.IOrderRepo {
	return &OrderRepository{db: db}
}

// orderEntity represents the database entity
type orderEntity struct {
	ID        string            `gorm:"primaryKey;type:uuid"`
	UserID    string            `gorm:"type:uuid;not null;index"`
	Total     float64           `gorm:"type:decimal(10,2);not null;default:0"`
	Status    string            `gorm:"not null;default:'pending'"`
	CreatedAt time.Time         `gorm:"autoCreateTime"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`
	DeletedAt *time.Time        `gorm:"index"`
	Items     []orderItemEntity `gorm:"foreignKey:OrderID"`
}

func (orderEntity) TableName() string {
	return "orders"
}

// orderItemEntity represents the order item database entity
type orderItemEntity struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	OrderID   string    `gorm:"type:uuid;not null;index"`
	ProductID string    `gorm:"not null;index"`
	Quantity  int       `gorm:"not null;default:1"`
	Price     float64   `gorm:"type:decimal(10,2);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (orderItemEntity) TableName() string {
	return "order_items"
}

// toModel converts entity to domain model
func (e *orderEntity) toModel() *model.Order {
	items := make([]model.OrderItem, len(e.Items))
	for i, item := range e.Items {
		items[i] = model.OrderItem{
			ID:        item.ID,
			OrderID:   item.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			CreatedAt: item.CreatedAt,
		}
	}

	return &model.Order{
		ID:        e.ID,
		UserID:    e.UserID,
		Items:     items,
		Total:     e.Total,
		Status:    model.OrderStatus(e.Status),
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: e.DeletedAt,
	}
}

// toEntity converts domain model to entity
func toOrderEntity(o *model.Order) *orderEntity {
	items := make([]orderItemEntity, len(o.Items))
	for i, item := range o.Items {
		items[i] = orderItemEntity{
			ID:        item.ID,
			OrderID:   item.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			CreatedAt: item.CreatedAt,
		}
	}

	return &orderEntity{
		ID:        o.ID,
		UserID:    o.UserID,
		Total:     o.Total,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		DeletedAt: o.DeletedAt,
		Items:     items,
	}
}

func (r *OrderRepository) getDB(ctx context.Context, tx repo.Transaction) *gorm.DB {
	if tx != nil {
		if gormTx, ok := tx.GetTx().(*gorm.DB); ok {
			return gormTx.WithContext(ctx)
		}
	}
	return r.db.WithContext(ctx)
}

// Create creates a new order with items
func (r *OrderRepository) Create(ctx context.Context, tx repo.Transaction, order *model.Order) (*model.Order, error) {
	entity := toOrderEntity(order)
	db := r.getDB(ctx, tx)

	if err := db.Create(entity).Error; err != nil {
		return nil, err
	}

	return entity.toModel(), nil
}

// Update updates an existing order
func (r *OrderRepository) Update(ctx context.Context, tx repo.Transaction, order *model.Order) error {
	entity := toOrderEntity(order)
	db := r.getDB(ctx, tx)

	return db.Save(entity).Error
}

// Delete soft deletes an order by ID
func (r *OrderRepository) Delete(ctx context.Context, tx repo.Transaction, id string) error {
	db := r.getDB(ctx, tx)

	now := time.Now()
	return db.Model(&orderEntity{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// GetByID retrieves an order by ID with items
func (r *OrderRepository) GetByID(ctx context.Context, tx repo.Transaction, id string) (*model.Order, error) {
	var entity orderEntity
	db := r.getDB(ctx, tx)

	err := db.Preload("Items").Where("id = ? AND deleted_at IS NULL", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return entity.toModel(), nil
}

// GetByUserID retrieves orders for a user with pagination
func (r *OrderRepository) GetByUserID(ctx context.Context, tx repo.Transaction, userID string, offset, limit int) ([]*model.Order, int64, error) {
	var entities []orderEntity
	var total int64
	db := r.getDB(ctx, tx)

	// Get total count
	if err := db.Model(&orderEntity{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with items
	if err := db.Preload("Items").Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	orders := make([]*model.Order, len(entities))
	for i, e := range entities {
		orders[i] = e.toModel()
	}

	return orders, total, nil
}

// List retrieves orders with pagination
func (r *OrderRepository) List(ctx context.Context, tx repo.Transaction, offset, limit int) ([]*model.Order, int64, error) {
	var entities []orderEntity
	var total int64
	db := r.getDB(ctx, tx)

	// Get total count
	if err := db.Model(&orderEntity{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with items
	if err := db.Preload("Items").Where("deleted_at IS NULL").
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	orders := make([]*model.Order, len(entities))
	for i, e := range entities {
		orders[i] = e.toModel()
	}

	return orders, total, nil
}

// UpdateStatus updates the order status
func (r *OrderRepository) UpdateStatus(ctx context.Context, tx repo.Transaction, id string, status model.OrderStatus) error {
	db := r.getDB(ctx, tx)

	return db.Model(&orderEntity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now(),
	}).Error
}
