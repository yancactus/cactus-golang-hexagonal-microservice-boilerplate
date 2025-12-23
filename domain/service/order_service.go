package service

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/event"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// IOrderService defines the interface for order service operations
type IOrderService interface {
	Create(ctx context.Context, userID string, items []model.OrderItem) (*model.Order, error)
	Get(ctx context.Context, id string) (*model.Order, error)
	GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*model.Order, int64, error)
	List(ctx context.Context, offset, limit int) ([]*model.Order, int64, error)
	UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error
	Cancel(ctx context.Context, id string) error
}

// OrderService implements IOrderService
type OrderService struct {
	repo      repo.IOrderRepo
	userRepo  repo.IUserRepo
	txFactory repo.TransactionFactory
	eventBus  event.EventBus
}

// NewOrderService creates a new order service
func NewOrderService(repo repo.IOrderRepo, userRepo repo.IUserRepo, txFactory repo.TransactionFactory, eventBus event.EventBus) *OrderService {
	if eventBus == nil {
		eventBus = event.NewNoopEventBus()
	}
	return &OrderService{
		repo:      repo,
		userRepo:  userRepo,
		txFactory: txFactory,
		eventBus:  eventBus,
	}
}

// Create creates a new order
func (s *OrderService) Create(ctx context.Context, userID string, items []model.OrderItem) (*model.Order, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, nil, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, model.ErrUserNotFound
	}

	// Create order
	order, err := model.NewOrder(userID, items)
	if err != nil {
		return nil, err
	}

	// Save to repository
	created, err := s.repo.Create(ctx, nil, order)
	if err != nil {
		return nil, err
	}

	// Publish domain events from original order (has the recorded events)
	s.publishEvents(ctx, order)

	return created, nil
}

// Get retrieves an order by ID
func (s *OrderService) Get(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.GetByID(ctx, nil, id)
}

// GetByUserID retrieves orders for a user
func (s *OrderService) GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*model.Order, int64, error) {
	return s.repo.GetByUserID(ctx, nil, userID, offset, limit)
}

// List retrieves orders with pagination
func (s *OrderService) List(ctx context.Context, offset, limit int) ([]*model.Order, int64, error) {
	return s.repo.List(ctx, nil, offset, limit)
}

// UpdateStatus updates the order status
func (s *OrderService) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error {
	order, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return err
	}
	if order == nil {
		return model.ErrOrderNotFound
	}

	// Validate status transition
	switch status {
	case model.OrderStatusConfirmed:
		if err := order.Confirm(); err != nil {
			return err
		}
	case model.OrderStatusShipped:
		if err := order.Ship(); err != nil {
			return err
		}
	case model.OrderStatusDelivered:
		if err := order.Deliver(); err != nil {
			return err
		}
	case model.OrderStatusCancelled:
		if err := order.Cancel(); err != nil {
			return err
		}
	default:
		return model.ErrOrderInvalidStatus
	}

	if err := s.repo.UpdateStatus(ctx, nil, id, status); err != nil {
		return err
	}

	// Publish domain events
	s.publishEvents(ctx, order)

	return nil
}

// Cancel cancels an order
func (s *OrderService) Cancel(ctx context.Context, id string) error {
	order, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return err
	}
	if order == nil {
		return model.ErrOrderNotFound
	}

	if err := order.Cancel(); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, nil, id, model.OrderStatusCancelled); err != nil {
		return err
	}

	// Publish domain events
	s.publishEvents(ctx, order)

	return nil
}

// publishEvents publishes all pending domain events from the order
func (s *OrderService) publishEvents(ctx context.Context, order *model.Order) {
	for _, domainEvent := range order.Events() {
		evt := event.NewBaseEvent(
			domainEvent.EventName(),
			order.ID,
			domainEvent,
		)
		if err := s.eventBus.Publish(ctx, evt); err != nil {
			log.SugaredLogger.Errorf("Failed to publish event %s: %v", domainEvent.EventName(), err)
		}
	}
}
