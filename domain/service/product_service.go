package service

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/event"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// IProductService defines the interface for product service operations
type IProductService interface {
	Create(ctx context.Context, name, description string, price float64, stock int) (*model.Product, error)
	Update(ctx context.Context, id string, name, description string, price float64) (*model.Product, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Product, error)
	GetByName(ctx context.Context, name string) (*model.Product, error)
	List(ctx context.Context, offset, limit int) ([]*model.Product, int64, error)
	UpdateStock(ctx context.Context, id string, quantity int) error
}

// ProductService implements IProductService
type ProductService struct {
	repo     repo.IProductRepo
	eventBus event.EventBus
}

// NewProductService creates a new product service
func NewProductService(repo repo.IProductRepo, eventBus event.EventBus) *ProductService {
	if eventBus == nil {
		eventBus = event.NewNoopEventBus()
	}
	return &ProductService{
		repo:     repo,
		eventBus: eventBus,
	}
}

// Create creates a new product
func (s *ProductService) Create(ctx context.Context, name, description string, price float64, stock int) (*model.Product, error) {
	product, err := model.NewProduct(name, description, price, stock)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	// Publish domain events from original product (has the recorded events)
	s.publishEvents(ctx, product)

	return created, nil
}

// Update updates an existing product
func (s *ProductService) Update(ctx context.Context, id string, name, description string, price float64) (*model.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, model.ErrProductNotFound
	}

	if err := product.Update(name, description, price); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	// Publish domain events
	s.publishEvents(ctx, product)

	return product, nil
}

// Delete deletes a product
func (s *ProductService) Delete(ctx context.Context, id string) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return model.ErrProductNotFound
	}

	// Mark as deleted (records event)
	product.MarkDeleted()

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish domain events
	s.publishEvents(ctx, product)

	return nil
}

// Get retrieves a product by ID
func (s *ProductService) Get(ctx context.Context, id string) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByName retrieves a product by name
func (s *ProductService) GetByName(ctx context.Context, name string) (*model.Product, error) {
	return s.repo.GetByName(ctx, name)
}

// List retrieves products with pagination
func (s *ProductService) List(ctx context.Context, offset, limit int) ([]*model.Product, int64, error) {
	return s.repo.List(ctx, offset, limit)
}

// UpdateStock updates product stock
func (s *ProductService) UpdateStock(ctx context.Context, id string, quantity int) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return model.ErrProductNotFound
	}

	if err := product.UpdateStock(quantity); err != nil {
		return err
	}

	if err := s.repo.UpdateStock(ctx, id, quantity); err != nil {
		return err
	}

	// Publish domain events
	s.publishEvents(ctx, product)

	return nil
}

// publishEvents publishes all pending domain events from the product
func (s *ProductService) publishEvents(ctx context.Context, product *model.Product) {
	for _, domainEvent := range product.Events() {
		evt := event.NewBaseEvent(
			domainEvent.EventName(),
			product.ID,
			domainEvent,
		)
		if err := s.eventBus.Publish(ctx, evt); err != nil {
			log.SugaredLogger.Errorf("Failed to publish event %s: %v", domainEvent.EventName(), err)
		}
	}
}
