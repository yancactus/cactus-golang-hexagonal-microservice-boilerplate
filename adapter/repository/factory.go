package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
)

// TransactionFactoryImpl implements the repo.TransactionFactory interface
type TransactionFactoryImpl struct {
	// Map of store types to clients
	clients map[StoreType]any
}

// NewTransactionFactory creates a new transaction factory with the provided clients
func NewTransactionFactory(clients map[StoreType]any) repo.TransactionFactory {
	return &TransactionFactoryImpl{
		clients: clients,
	}
}

// NewTransaction creates a new transaction for the specified store
func (f *TransactionFactoryImpl) NewTransaction(ctx context.Context, store repo.StoreType, opts any) (repo.Transaction, error) {
	// Convert domain StoreType to adapter StoreType
	adapterStore := StoreType(store)

	// Get the client for the store type
	client, ok := f.clients[adapterStore]
	if !ok {
		return nil, fmt.Errorf("no client found for store type: %s", store)
	}

	// Convert options to SQL options if applicable
	var sqlOpts *sql.TxOptions
	if opt, ok := opts.(*sql.TxOptions); ok {
		sqlOpts = opt
	}

	// Create transaction
	tx, err := NewTransaction(ctx, adapterStore, client, sqlOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

// OpenPostgresGormDB creates a new GORM database connection based on the PostgreSQL configuration
func OpenPostgresGormDB() (*gorm.DB, error) {
	if config.GlobalConfig.Postgre == nil {
		return nil, ErrMissingPostgreSQLConfig
	}

	// Construct DSN from configuration
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.GlobalConfig.Postgre.Host,
		config.GlobalConfig.Postgre.User,
		config.GlobalConfig.Postgre.Password,
		config.GlobalConfig.Postgre.Database,
		config.GlobalConfig.Postgre.Port,
		config.GlobalConfig.Postgre.SSLMode,
		config.GlobalConfig.Postgre.TimeZone,
	)

	// Configure GORM
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB: %w", err)
	}

	if config.GlobalConfig.Postgre.MaxConnections > 0 {
		sqlDB.SetMaxOpenConns(int(config.GlobalConfig.Postgre.MaxConnections))
	}
	if config.GlobalConfig.Postgre.MinConnections > 0 {
		sqlDB.SetMaxIdleConns(int(config.GlobalConfig.Postgre.MinConnections))
	}
	if config.GlobalConfig.Postgre.MaxConnLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(config.GlobalConfig.Postgre.MaxConnLifetime) * time.Second)
	}
	if config.GlobalConfig.Postgre.IdleTimeout > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(config.GlobalConfig.Postgre.IdleTimeout) * time.Second)
	}

	return db, nil
}

// NewRedisConn creates a new Redis client connection based on the Redis configuration
func NewRedisConn() *redis.Client {
	if config.GlobalConfig.Redis == nil {
		return nil
	}

	// Create Redis client from configuration
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.GlobalConfig.Redis.Host, config.GlobalConfig.Redis.Port),
		Password:     config.GlobalConfig.Redis.Password,
		DB:           config.GlobalConfig.Redis.DB,
		PoolSize:     config.GlobalConfig.Redis.PoolSize,
		MinIdleConns: config.GlobalConfig.Redis.MinIdleConns,
		IdleTimeout:  time.Duration(config.GlobalConfig.Redis.IdleTimeout) * time.Second,
	})

	return client
}

// NewPostgreConn creates a new PostgreSQL connection pool based on the PostgreSQL configuration
func NewPostgreConn() (*pgxpool.Pool, error) {
	if config.GlobalConfig.Postgre == nil {
		return nil, ErrMissingPostgreSQLConfig
	}

	// Construct connection string
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.GlobalConfig.Postgre.User,
		config.GlobalConfig.Postgre.Password,
		config.GlobalConfig.Postgre.Host,
		config.GlobalConfig.Postgre.Port,
		config.GlobalConfig.Postgre.Database,
		config.GlobalConfig.Postgre.SSLMode,
	)

	// Parse connection config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL connection string: %w", err)
	}

	// Configure connection pool
	if config.GlobalConfig.Postgre.MaxConnections > 0 {
		poolConfig.MaxConns = config.GlobalConfig.Postgre.MaxConnections
	}
	if config.GlobalConfig.Postgre.MinConnections > 0 {
		poolConfig.MinConns = config.GlobalConfig.Postgre.MinConnections
	}
	if config.GlobalConfig.Postgre.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = time.Duration(config.GlobalConfig.Postgre.MaxConnLifetime) * time.Second
	}
	if config.GlobalConfig.Postgre.IdleTimeout > 0 {
		poolConfig.MaxConnIdleTime = time.Duration(config.GlobalConfig.Postgre.IdleTimeout) * time.Second
	}
	if config.GlobalConfig.Postgre.ConnectTimeout > 0 {
		poolConfig.ConnConfig.ConnectTimeout = time.Duration(config.GlobalConfig.Postgre.ConnectTimeout) * time.Second
	}

	// Create connection pool
	return pgxpool.NewWithConfig(context.Background(), poolConfig)
}
