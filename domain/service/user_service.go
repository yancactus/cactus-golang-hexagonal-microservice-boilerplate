package service

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/event"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

const userServiceTracerName = "user-service"

// IUserService defines the interface for user service operations
type IUserService interface {
	Create(ctx context.Context, email, name, password string) (*model.User, error)
	Update(ctx context.Context, id string, name string) (*model.User, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, offset, limit int) ([]*model.User, int64, error)
}

// UserService implements IUserService
type UserService struct {
	repo      repo.IUserRepo
	txFactory repo.TransactionFactory
	eventBus  event.EventBus
}

// NewUserService creates a new user service
func NewUserService(repo repo.IUserRepo, txFactory repo.TransactionFactory, eventBus event.EventBus) *UserService {
	if eventBus == nil {
		eventBus = event.NewNoopEventBus()
	}
	return &UserService{
		repo:      repo,
		txFactory: txFactory,
		eventBus:  eventBus,
	}
}

// Create creates a new user
func (s *UserService) Create(ctx context.Context, email, name, password string) (*model.User, error) {
	ctx, span := otel.Tracer(userServiceTracerName).Start(ctx, "UserService.Create")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	// Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, nil, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	if existing != nil {
		span.SetStatus(codes.Error, "email already taken")
		return nil, model.ErrUserEmailTaken
	}

	// Create user
	user, err := model.NewUser(email, name, password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Save to repository
	created, err := s.repo.Create(ctx, nil, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.String("user.id", created.ID))
	span.SetStatus(codes.Ok, "user created")

	// Publish domain events from original user (has the recorded events)
	s.publishEvents(ctx, user)

	return created, nil
}

// Update updates an existing user
func (s *UserService) Update(ctx context.Context, id string, name string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, model.ErrUserNotFound
	}

	if err := user.Update(name); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, nil, user); err != nil {
		return nil, err
	}

	// Publish domain events
	s.publishEvents(ctx, user)

	return user, nil
}

// Delete deletes a user
func (s *UserService) Delete(ctx context.Context, id string) error {
	user, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return err
	}
	if user == nil {
		return model.ErrUserNotFound
	}

	// Mark as deleted (records event)
	user.MarkDeleted()

	if err := s.repo.Delete(ctx, nil, id); err != nil {
		return err
	}

	// Publish domain events
	s.publishEvents(ctx, user)

	return nil
}

// Get retrieves a user by ID
func (s *UserService) Get(ctx context.Context, id string) (*model.User, error) {
	ctx, span := otel.Tracer(userServiceTracerName).Start(ctx, "UserService.Get")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id))

	user, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if user != nil {
		span.SetStatus(codes.Ok, "user found")
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.repo.GetByEmail(ctx, nil, email)
}

// List retrieves users with pagination
func (s *UserService) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	return s.repo.List(ctx, nil, offset, limit)
}

// publishEvents publishes all pending domain events from the user
func (s *UserService) publishEvents(ctx context.Context, user *model.User) {
	for _, domainEvent := range user.Events() {
		evt := event.NewBaseEvent(
			domainEvent.EventName(),
			user.ID,
			domainEvent,
		)
		if err := s.eventBus.Publish(ctx, evt); err != nil {
			log.SugaredLogger.Errorf("Failed to publish event %s: %v", domainEvent.EventName(), err)
		}
	}
}
