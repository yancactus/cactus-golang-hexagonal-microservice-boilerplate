package service

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository/redis"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
	"cactus-golang-hexagonal-microservice-boilerplate/util/metrics"
)

const (
	cachedUserServiceTracerName = "cached-user-service"
	userCacheKeyPrefix          = "user:"
	userEmailCacheKeyPrefix     = "user:email:"
	defaultUserCacheTTL         = 30 * time.Minute
)

// CachedUserService wraps a UserService with Redis caching
type CachedUserService struct {
	delegate IUserService
	cache    *redis.EnhancedCache
	ttl      time.Duration
}

// NewCachedUserService creates a new cached user service
func NewCachedUserService(delegate IUserService, cache *redis.EnhancedCache) *CachedUserService {
	return &CachedUserService{
		delegate: delegate,
		cache:    cache,
		ttl:      defaultUserCacheTTL,
	}
}

// Create creates a new user and caches the result
func (s *CachedUserService) Create(ctx context.Context, email, name, password string) (*model.User, error) {
	ctx, span := otel.Tracer(cachedUserServiceTracerName).Start(ctx, "CachedUserService.Create")
	defer span.End()

	// Delegate to the underlying service
	user, err := s.delegate.Create(ctx, email, name, password)
	if err != nil {
		return nil, err
	}

	// Cache the new user
	s.cacheUser(ctx, user)

	return user, nil
}

// Update updates a user and invalidates the cache
func (s *CachedUserService) Update(ctx context.Context, id string, name string) (*model.User, error) {
	ctx, span := otel.Tracer(cachedUserServiceTracerName).Start(ctx, "CachedUserService.Update")
	defer span.End()

	// Get the current user to invalidate email cache
	currentUser, _ := s.delegate.Get(ctx, id)

	// Delegate to the underlying service
	user, err := s.delegate.Update(ctx, id, name)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	s.invalidateUserCache(ctx, id)
	if currentUser != nil {
		s.invalidateEmailCache(ctx, currentUser.Email)
	}

	// Cache the updated user
	s.cacheUser(ctx, user)

	return user, nil
}

// Delete deletes a user and invalidates the cache
func (s *CachedUserService) Delete(ctx context.Context, id string) error {
	ctx, span := otel.Tracer(cachedUserServiceTracerName).Start(ctx, "CachedUserService.Delete")
	defer span.End()

	// Get the current user to invalidate email cache
	currentUser, _ := s.delegate.Get(ctx, id)

	// Delegate to the underlying service
	err := s.delegate.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateUserCache(ctx, id)
	if currentUser != nil {
		s.invalidateEmailCache(ctx, currentUser.Email)
	}

	return nil
}

// Get retrieves a user by ID, using cache when available
func (s *CachedUserService) Get(ctx context.Context, id string) (*model.User, error) {
	ctx, span := otel.Tracer(cachedUserServiceTracerName).Start(ctx, "CachedUserService.Get")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id))

	// Try cache first
	cacheKey := s.userCacheKey(id)
	var user model.User

	err := s.cache.Get(ctx, cacheKey, &user)
	if err == nil {
		// Cache hit
		span.SetAttributes(attribute.Bool("cache.hit", true))
		metrics.RecordCacheHit("user", "hit")
		log.SugaredLogger.Debugf("Cache hit for user %s", id)
		return &user, nil
	}

	// Cache miss - get from delegate
	span.SetAttributes(attribute.Bool("cache.hit", false))
	metrics.RecordCacheHit("user", "miss")

	result, err := s.delegate.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result if found
	if result != nil {
		s.cacheUser(ctx, result)
	}

	return result, nil
}

// GetByEmail retrieves a user by email, using cache when available
func (s *CachedUserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := otel.Tracer(cachedUserServiceTracerName).Start(ctx, "CachedUserService.GetByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	// Try email->id cache first
	emailCacheKey := s.emailCacheKey(email)
	var userID string

	err := s.cache.Get(ctx, emailCacheKey, &userID)
	if err == nil && userID != "" {
		// Found user ID in cache, now get user by ID (which also uses cache)
		span.SetAttributes(attribute.Bool("email_cache.hit", true))
		return s.Get(ctx, userID)
	}

	span.SetAttributes(attribute.Bool("email_cache.hit", false))

	// Cache miss - get from delegate
	result, err := s.delegate.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Cache the result if found
	if result != nil {
		s.cacheUser(ctx, result)
		// Also cache email -> id mapping
		if cacheErr := s.cache.Set(ctx, emailCacheKey, result.ID, s.ttl); cacheErr != nil {
			log.SugaredLogger.Warnf("Failed to cache email mapping for %s: %v", email, cacheErr)
		}
	}

	return result, nil
}

// List retrieves users with pagination (not cached - lists are dynamic)
func (s *CachedUserService) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	// Lists are not cached as they are dynamic
	return s.delegate.List(ctx, offset, limit)
}

// Helper methods

func (s *CachedUserService) userCacheKey(id string) string {
	return fmt.Sprintf("%s%s", userCacheKeyPrefix, id)
}

func (s *CachedUserService) emailCacheKey(email string) string {
	return fmt.Sprintf("%s%s", userEmailCacheKeyPrefix, email)
}

func (s *CachedUserService) cacheUser(ctx context.Context, user *model.User) {
	if user == nil {
		return
	}

	cacheKey := s.userCacheKey(user.ID)
	if err := s.cache.Set(ctx, cacheKey, user, s.ttl); err != nil {
		log.SugaredLogger.Warnf("Failed to cache user %s: %v", user.ID, err)
	}

	// Also cache email -> id mapping
	emailCacheKey := s.emailCacheKey(user.Email)
	if err := s.cache.Set(ctx, emailCacheKey, user.ID, s.ttl); err != nil {
		log.SugaredLogger.Warnf("Failed to cache email mapping for %s: %v", user.Email, err)
	}
}

func (s *CachedUserService) invalidateUserCache(ctx context.Context, id string) {
	cacheKey := s.userCacheKey(id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.SugaredLogger.Warnf("Failed to invalidate user cache %s: %v", id, err)
	}
}

func (s *CachedUserService) invalidateEmailCache(ctx context.Context, email string) {
	cacheKey := s.emailCacheKey(email)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.SugaredLogger.Warnf("Failed to invalidate email cache %s: %v", email, err)
	}
}
