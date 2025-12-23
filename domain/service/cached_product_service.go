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
	cachedProductServiceTracerName = "cached-product-service"
	productCacheKeyPrefix          = "product:"
	productNameCacheKeyPrefix      = "product:name:"
	defaultProductCacheTTL         = 30 * time.Minute
)

// CachedProductService wraps a ProductService with Redis caching
type CachedProductService struct {
	delegate IProductService
	cache    *redis.EnhancedCache
	ttl      time.Duration
}

// NewCachedProductService creates a new cached product service
func NewCachedProductService(delegate IProductService, cache *redis.EnhancedCache) *CachedProductService {
	return &CachedProductService{
		delegate: delegate,
		cache:    cache,
		ttl:      defaultProductCacheTTL,
	}
}

// Create creates a new product and caches the result
func (s *CachedProductService) Create(ctx context.Context, name, description string, price float64, stock int) (*model.Product, error) {
	ctx, span := otel.Tracer(cachedProductServiceTracerName).Start(ctx, "CachedProductService.Create")
	defer span.End()

	// Delegate to the underlying service
	product, err := s.delegate.Create(ctx, name, description, price, stock)
	if err != nil {
		return nil, err
	}

	// Cache the new product
	s.cacheProduct(ctx, product)

	return product, nil
}

// Update updates a product and invalidates the cache
func (s *CachedProductService) Update(ctx context.Context, id string, name, description string, price float64) (*model.Product, error) {
	ctx, span := otel.Tracer(cachedProductServiceTracerName).Start(ctx, "CachedProductService.Update")
	defer span.End()

	// Get the current product to invalidate name cache
	currentProduct, _ := s.delegate.Get(ctx, id)

	// Delegate to the underlying service
	product, err := s.delegate.Update(ctx, id, name, description, price)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	s.invalidateProductCache(ctx, id)
	if currentProduct != nil {
		s.invalidateNameCache(ctx, currentProduct.Name)
	}

	// Cache the updated product
	s.cacheProduct(ctx, product)

	return product, nil
}

// Delete deletes a product and invalidates the cache
func (s *CachedProductService) Delete(ctx context.Context, id string) error {
	ctx, span := otel.Tracer(cachedProductServiceTracerName).Start(ctx, "CachedProductService.Delete")
	defer span.End()

	// Get the current product to invalidate name cache
	currentProduct, _ := s.delegate.Get(ctx, id)

	// Delegate to the underlying service
	err := s.delegate.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateProductCache(ctx, id)
	if currentProduct != nil {
		s.invalidateNameCache(ctx, currentProduct.Name)
	}

	return nil
}

// Get retrieves a product by ID, using cache when available
func (s *CachedProductService) Get(ctx context.Context, id string) (*model.Product, error) {
	ctx, span := otel.Tracer(cachedProductServiceTracerName).Start(ctx, "CachedProductService.Get")
	defer span.End()

	span.SetAttributes(attribute.String("product.id", id))

	// Try cache first
	cacheKey := s.productCacheKey(id)
	var product model.Product

	err := s.cache.Get(ctx, cacheKey, &product)
	if err == nil {
		// Cache hit
		span.SetAttributes(attribute.Bool("cache.hit", true))
		metrics.RecordCacheHit("product", "hit")
		log.SugaredLogger.Debugf("Cache hit for product %s", id)
		return &product, nil
	}

	// Cache miss - get from delegate
	span.SetAttributes(attribute.Bool("cache.hit", false))
	metrics.RecordCacheHit("product", "miss")

	result, err := s.delegate.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result if found
	if result != nil {
		s.cacheProduct(ctx, result)
	}

	return result, nil
}

// GetByName retrieves a product by name, using cache when available
func (s *CachedProductService) GetByName(ctx context.Context, name string) (*model.Product, error) {
	ctx, span := otel.Tracer(cachedProductServiceTracerName).Start(ctx, "CachedProductService.GetByName")
	defer span.End()

	span.SetAttributes(attribute.String("product.name", name))

	// Try name->id cache first
	nameCacheKey := s.nameCacheKey(name)
	var productID string

	err := s.cache.Get(ctx, nameCacheKey, &productID)
	if err == nil && productID != "" {
		// Found product ID in cache, now get product by ID (which also uses cache)
		span.SetAttributes(attribute.Bool("name_cache.hit", true))
		return s.Get(ctx, productID)
	}

	span.SetAttributes(attribute.Bool("name_cache.hit", false))

	// Cache miss - get from delegate
	result, err := s.delegate.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// Cache the result if found
	if result != nil {
		s.cacheProduct(ctx, result)
		// Also cache name -> id mapping
		if cacheErr := s.cache.Set(ctx, nameCacheKey, result.ID, s.ttl); cacheErr != nil {
			log.SugaredLogger.Warnf("Failed to cache name mapping for %s: %v", name, cacheErr)
		}
	}

	return result, nil
}

// List retrieves products with pagination (not cached - lists are dynamic)
func (s *CachedProductService) List(ctx context.Context, offset, limit int) ([]*model.Product, int64, error) {
	// Lists are not cached as they are dynamic
	return s.delegate.List(ctx, offset, limit)
}

// UpdateStock updates product stock and invalidates the cache
func (s *CachedProductService) UpdateStock(ctx context.Context, id string, quantity int) error {
	ctx, span := otel.Tracer(cachedProductServiceTracerName).Start(ctx, "CachedProductService.UpdateStock")
	defer span.End()

	// Delegate to the underlying service
	err := s.delegate.UpdateStock(ctx, id, quantity)
	if err != nil {
		return err
	}

	// Invalidate cache (stock changed)
	s.invalidateProductCache(ctx, id)

	return nil
}

// Helper methods

func (s *CachedProductService) productCacheKey(id string) string {
	return fmt.Sprintf("%s%s", productCacheKeyPrefix, id)
}

func (s *CachedProductService) nameCacheKey(name string) string {
	return fmt.Sprintf("%s%s", productNameCacheKeyPrefix, name)
}

func (s *CachedProductService) cacheProduct(ctx context.Context, product *model.Product) {
	if product == nil {
		return
	}

	cacheKey := s.productCacheKey(product.ID)
	if err := s.cache.Set(ctx, cacheKey, product, s.ttl); err != nil {
		log.SugaredLogger.Warnf("Failed to cache product %s: %v", product.ID, err)
	}

	// Also cache name -> id mapping
	nameCacheKey := s.nameCacheKey(product.Name)
	if err := s.cache.Set(ctx, nameCacheKey, product.ID, s.ttl); err != nil {
		log.SugaredLogger.Warnf("Failed to cache name mapping for %s: %v", product.Name, err)
	}
}

func (s *CachedProductService) invalidateProductCache(ctx context.Context, id string) {
	cacheKey := s.productCacheKey(id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.SugaredLogger.Warnf("Failed to invalidate product cache %s: %v", id, err)
	}
}

func (s *CachedProductService) invalidateNameCache(ctx context.Context, name string) {
	cacheKey := s.nameCacheKey(name)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.SugaredLogger.Warnf("Failed to invalidate name cache %s: %v", name, err)
	}
}
