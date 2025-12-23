// Package dependency provides dependency injection configuration
//go:build wireinject
// +build wireinject

package dependency

import (
	"context"

	goredis "github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/gorm"

	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository/dynamodb"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository/mongo"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository/postgre"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository/redis"
	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/event"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// RepositoryOption defines an option for repository initialization
type RepositoryOption func(*repository.ClientContainer)

// WithPostgres returns an option to initialize PostgreSQL
func WithPostgres() RepositoryOption {
	return func(c *repository.ClientContainer) {
		if c.PostgreSQL == nil {
			pg, err := ProvidePostgres()
			if err != nil {
				panic("Failed to initialize PostgreSQL: " + err.Error())
			}
			c.PostgreSQL = pg
		}
	}
}

// WithMongoDB returns an option to initialize MongoDB
func WithMongoDB() RepositoryOption {
	return func(c *repository.ClientContainer) {
		if c.MongoDB == nil {
			mongodb, err := ProvideMongoDB()
			if err != nil {
				panic("Failed to initialize MongoDB: " + err.Error())
			}
			c.MongoDB = mongodb
		}
	}
}

// WithDynamoDB returns an option to initialize DynamoDB
func WithDynamoDB() RepositoryOption {
	return func(c *repository.ClientContainer) {
		if c.DynamoDB == nil {
			dynamo, err := ProvideDynamoDB()
			if err != nil {
				panic("Failed to initialize DynamoDB: " + err.Error())
			}
			c.DynamoDB = dynamo
		}
	}
}

// WithRedis returns an option to initialize Redis
func WithRedis() RepositoryOption {
	return func(c *repository.ClientContainer) {
		if c.Redis == nil {
			redisClient, err := ProvideRedis()
			if err != nil {
				panic("Failed to initialize Redis: " + err.Error())
			}
			c.Redis = redisClient
		}
	}
}

// ServiceOption defines an option for service initialization
type ServiceOption func(*service.Services, event.EventBus, *repository.ClientContainer)

// WithUserService returns an option to initialize the User service
func WithUserService() ServiceOption {
	return func(s *service.Services, eventBus event.EventBus, c *repository.ClientContainer) {
		if s.UserService == nil && c.PostgreSQL != nil {
			userRepo := postgre.NewUserRepository(c.PostgreSQL.DB)
			txFactory := repository.NewTransactionFactory(map[repository.StoreType]any{
				repository.PostgreSQLStore: c.PostgreSQL,
			})
			s.UserService = service.NewUserService(userRepo, txFactory, eventBus)
		}
	}
}

// WithProductService returns an option to initialize the Product service
func WithProductService() ServiceOption {
	return func(s *service.Services, eventBus event.EventBus, c *repository.ClientContainer) {
		if s.ProductService == nil && c.MongoDB != nil {
			if mongoClient, ok := c.MongoDB.Client.(*mongo.Client); ok {
				productRepo := mongo.NewProductRepository(mongoClient)
				s.ProductService = service.NewProductService(productRepo, eventBus)
			}
		}
	}
}

// WithOrderService returns an option to initialize the Order service
func WithOrderService() ServiceOption {
	return func(s *service.Services, eventBus event.EventBus, c *repository.ClientContainer) {
		if s.OrderService == nil && c.PostgreSQL != nil {
			orderRepo := postgre.NewOrderRepository(c.PostgreSQL.DB)
			userRepo := postgre.NewUserRepository(c.PostgreSQL.DB)
			txFactory := repository.NewTransactionFactory(map[repository.StoreType]any{
				repository.PostgreSQLStore: c.PostgreSQL,
			})
			s.OrderService = service.NewOrderService(orderRepo, userRepo, txFactory, eventBus)
		}
	}
}

// WithCachedUserService returns an option to initialize the User service with Redis caching
func WithCachedUserService() ServiceOption {
	return func(s *service.Services, eventBus event.EventBus, c *repository.ClientContainer) {
		if s.UserService == nil && c.PostgreSQL != nil && c.Redis != nil {
			// Create base user service
			userRepo := postgre.NewUserRepository(c.PostgreSQL.DB)
			txFactory := repository.NewTransactionFactory(map[repository.StoreType]any{
				repository.PostgreSQLStore: c.PostgreSQL,
			})
			baseService := service.NewUserService(userRepo, txFactory, eventBus)

			// Create Redis client and enhanced cache
			redisClient, err := redis.NewClientFromConfig(config.GlobalConfig.Redis)
			if err != nil {
				// Fall back to base service without caching
				s.UserService = baseService
				return
			}
			cache := redis.NewEnhancedCache(redisClient, redis.DefaultCacheOptions())

			// Wrap with caching
			s.UserService = service.NewCachedUserService(baseService, cache)
		}
	}
}

// WithCachedProductService returns an option to initialize the Product service with Redis caching
func WithCachedProductService() ServiceOption {
	return func(s *service.Services, eventBus event.EventBus, c *repository.ClientContainer) {
		if s.ProductService == nil && c.MongoDB != nil && c.Redis != nil {
			if mongoClient, ok := c.MongoDB.Client.(*mongo.Client); ok {
				// Create base product service
				productRepo := mongo.NewProductRepository(mongoClient)
				baseService := service.NewProductService(productRepo, eventBus)

				// Create Redis client and enhanced cache
				redisClient, err := redis.NewClientFromConfig(config.GlobalConfig.Redis)
				if err != nil {
					// Fall back to base service without caching
					s.ProductService = baseService
					return
				}
				cache := redis.NewEnhancedCache(redisClient, redis.DefaultCacheOptions())

				// Wrap with caching
				s.ProductService = service.NewCachedProductService(baseService, cache)
			}
		}
	}
}

// InitializeServices initializes services based on the provided options
func InitializeServices(ctx context.Context, clients *repository.ClientContainer, eventBus event.EventBus, opts ...ServiceOption) (*service.Services, error) {
	if eventBus == nil {
		eventBus = provideEventBus()
	}
	services := &service.Services{
		EventBus: eventBus,
	}

	for _, opt := range opts {
		opt(services, eventBus, clients)
	}

	return services, nil
}

// InitializeRepositories initializes repository clients with the given options
func InitializeRepositories(opts ...RepositoryOption) (*repository.ClientContainer, error) {
	container := &repository.ClientContainer{}
	for _, opt := range opts {
		opt(container)
	}
	return container, nil
}

// ProvidePostgres creates and initializes a PostgreSQL client
func ProvidePostgres() (*repository.PostgreSQL, error) {
	if config.GlobalConfig.Postgre == nil {
		return nil, repository.ErrMissingPostgreSQLConfig
	}

	db, err := repository.OpenPostgresGormDB()
	if err != nil {
		return nil, err
	}

	return &repository.PostgreSQL{DB: db}, nil
}

// ProvideMongoDB creates and initializes a MongoDB client
func ProvideMongoDB() (*repository.MongoDB, error) {
	if config.GlobalConfig.MongoDB == nil {
		return nil, repository.ErrMissingMongoDBConfig
	}

	client, err := mongo.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	return &repository.MongoDB{
		Client:   client,
		Database: config.GlobalConfig.MongoDB.Database,
	}, nil
}

// ProvideDynamoDB creates and initializes a DynamoDB client
func ProvideDynamoDB() (*repository.DynamoDB, error) {
	if config.GlobalConfig.DynamoDB == nil {
		return nil, repository.ErrMissingDynamoDBConfig
	}

	client, err := dynamodb.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	return &repository.DynamoDB{Client: client}, nil
}

// ProvideRedis creates and initializes a Redis client
func ProvideRedis() (*repository.Redis, error) {
	if config.GlobalConfig.Redis == nil {
		return nil, repository.ErrMissingRedisConfig
	}

	client := repository.NewRedisConn()
	return &repository.Redis{DB: client}, nil
}

// ProvideTransactionFactory creates and initializes a transaction factory
func ProvideTransactionFactory(clients map[repository.StoreType]any) repo.TransactionFactory {
	return repository.NewTransactionFactory(clients)
}

// PostgresSet provides a Wire provider set for PostgreSQL
var PostgresSet = wire.NewSet(
	ProvidePostgres,
	wire.Bind(new(PostgresRepository), new(*repository.PostgreSQL)),
)

// MongoDBSet provides a Wire provider set for MongoDB
var MongoDBSet = wire.NewSet(
	ProvideMongoDB,
	wire.Bind(new(MongoDBRepository), new(*repository.MongoDB)),
)

// DynamoDBSet provides a Wire provider set for DynamoDB
var DynamoDBSet = wire.NewSet(
	ProvideDynamoDB,
	wire.Bind(new(DynamoDBRepository), new(*repository.DynamoDB)),
)

// RedisSet provides a Wire provider set for Redis
var RedisSet = wire.NewSet(
	ProvideRedis,
	wire.Bind(new(RedisRepository), new(*repository.Redis)),
)

// PostgresRepository defines the interface for PostgreSQL operations
type PostgresRepository interface {
	GetDB(ctx context.Context) *gorm.DB
	Close(ctx context.Context) error
}

// MongoDBRepository defines the interface for MongoDB operations
type MongoDBRepository interface {
	Close(ctx context.Context) error
}

// DynamoDBRepository defines the interface for DynamoDB operations
type DynamoDBRepository interface {
	Close(ctx context.Context) error
}

// RedisRepository defines the interface for Redis operations
type RedisRepository interface {
	GetClient() *goredis.Client
	Close(ctx context.Context) error
}

// provideEventBus creates and configures the event bus
func provideEventBus() *event.InMemoryEventBus {
	eventBus := event.NewInMemoryEventBus()

	// Register event handlers
	loggingHandler := event.NewLoggingEventHandler()
	eventBus.Subscribe(loggingHandler)

	return eventBus
}
