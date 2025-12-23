package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cactus-golang-hexagonal-microservice-boilerplate/adapter/amqp"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/dependency"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/job"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository/dynamodb"
	"cactus-golang-hexagonal-microservice-boilerplate/api/middleware"
	"cactus-golang-hexagonal-microservice-boilerplate/cmd/http_server"
	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/event"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
	"cactus-golang-hexagonal-microservice-boilerplate/util/tracing"

	"go.uber.org/zap"
)

const ServiceName = "cactus-golang-hexagonal-microservice-boilerplate"

// Constants for application settings
const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 5 * time.Second
	// DefaultMetricsAddr is the default address for the metrics server
	DefaultMetricsAddr = ":9090"
)

func main() {
	fmt.Println("Starting " + ServiceName)

	config.Init("./config", "config")
	fmt.Println("Configuration initialized")

	log.Init()
	log.Logger.Info("Application starting",
		zap.String("service", ServiceName),
		zap.String("env", string(config.GlobalConfig.Env)))

	middleware.InitializeMetrics()
	log.Logger.Info("Metrics collection system initialized")

	// Initialize OpenTelemetry tracing
	if config.GlobalConfig.Tracing != nil && config.GlobalConfig.Tracing.Enabled {
		tracingCfg := &tracing.Config{
			Enabled:     true,
			Endpoint:    config.GlobalConfig.Tracing.Endpoint,
			ServiceName: ServiceName,
			Environment: string(config.GlobalConfig.Env),
			Sampler:     config.GlobalConfig.Tracing.Sampler,
		}
		tp, err := tracing.Init(context.Background(), tracingCfg)
		if err != nil {
			log.Logger.Warn("Failed to initialize tracing", zap.Error(err))
		} else {
			defer func() {
				if err := tp.Shutdown(context.Background()); err != nil {
					log.Logger.Error("Failed to shutdown tracer provider", zap.Error(err))
				}
			}()
			log.Logger.Info("OpenTelemetry tracing initialized",
				zap.String("endpoint", config.GlobalConfig.Tracing.Endpoint))
		}
	} else {
		log.Logger.Info("OpenTelemetry tracing is disabled")
	}

	if config.GlobalConfig.MetricsServer != nil && config.GlobalConfig.MetricsServer.Enabled {
		metricsAddr := config.GlobalConfig.MetricsServer.Addr
		if metricsAddr == "" {
			metricsAddr = DefaultMetricsAddr
		}
		go func() {
			if err := middleware.StartMetricsServer(metricsAddr); err != nil {
				log.Logger.Error("Failed to start metrics server", zap.Error(err))
			}
		}()
		log.Logger.Info("Metrics server started", zap.String("address", metricsAddr))
	} else {
		log.Logger.Info("Metrics server is disabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Logger.Info("Initializing repositories")
	clients, err := dependency.InitializeRepositories(
		dependency.WithPostgres(),
		dependency.WithMongoDB(),
		dependency.WithDynamoDB(),
		dependency.WithRedis(),
	)
	if err != nil {
		log.Logger.Fatal("Failed to initialize repositories",
			zap.Error(err))
	}
	repository.Clients = clients
	log.Logger.Info("Repositories initialized successfully")

	var kafkaProducer *amqp.KafkaEventBus
	var kafkaConsumer *amqp.KafkaConsumer
	var eventBus event.EventBus

	if config.GlobalConfig.Kafka != nil {
		log.Logger.Info("Initializing Kafka producer for audit events")

		kafkaCfg := &amqp.KafkaConfig{
			Brokers: config.GlobalConfig.Kafka.Brokers,
			Topic:   config.GlobalConfig.Kafka.Topics.AuditEvents,
		}
		var err error
		kafkaProducer, err = amqp.NewKafkaEventBus(kafkaCfg)
		if err != nil {
			log.Logger.Warn("Failed to initialize Kafka producer, using default event bus", zap.Error(err))
		} else {
			inMemoryBus := event.NewInMemoryEventBus()
			auditTopic := config.GlobalConfig.Kafka.Topics.AuditEvents
			kafkaAuditHandler := event.NewKafkaAuditHandler(kafkaProducer, auditTopic)
			inMemoryBus.Subscribe(kafkaAuditHandler)
			eventBus = inMemoryBus
			log.Logger.Info("Kafka audit handler registered")
		}
	} else {
		log.Logger.Info("Kafka not configured, using default event bus")
	}

	log.Logger.Info("Initializing services")
	// Use cached services if Redis is available, otherwise use regular services
	var serviceOpts []dependency.ServiceOption
	if clients.Redis != nil {
		log.Logger.Info("Redis available - using cached services")
		serviceOpts = []dependency.ServiceOption{
			dependency.WithCachedUserService(),
			dependency.WithCachedProductService(),
			dependency.WithOrderService(),
		}
	} else {
		log.Logger.Info("Redis not available - using regular services")
		serviceOpts = []dependency.ServiceOption{
			dependency.WithUserService(),
			dependency.WithProductService(),
			dependency.WithOrderService(),
		}
	}
	services, err := dependency.InitializeServices(ctx, clients, eventBus, serviceOpts...)
	if err != nil {
		log.Logger.Fatal("Failed to initialize services",
			zap.Error(err))
	}
	log.Logger.Info("Services initialized successfully")

	// Initialize DynamoDB and Audit consumer (requires DynamoDB)
	if config.GlobalConfig.DynamoDB != nil {
		log.Logger.Info("Initializing DynamoDB client for audit service")
		dynamoClient, err := dynamodb.NewClient(ctx)
		if err != nil {
			log.Logger.Warn("Failed to initialize DynamoDB client, audit service disabled", zap.Error(err))
		} else {
			// Initialize audit repository and service
			auditRepo := dynamodb.NewAuditRepository(dynamoClient)
			services.AuditService = service.NewAuditService(auditRepo)
			log.Logger.Info("Audit service initialized")

			// Initialize Kafka consumer if Kafka is available
			if kafkaProducer != nil {
				// Initialize audit consumer handler (uses AuditService)
				auditHandler := job.NewAuditConsumerHandler(services.AuditService)

				// Initialize Kafka consumer
				kafkaConsumer, err = amqp.NewKafkaConsumer(auditHandler)
				if err != nil {
					log.Logger.Warn("Failed to initialize Kafka consumer", zap.Error(err))
				} else {
					// Start Kafka consumer in background
					go func() {
						if err := kafkaConsumer.Start(); err != nil {
							log.Logger.Error("Kafka consumer failed", zap.Error(err))
						}
					}()
					log.Logger.Info("Kafka audit consumer started")
				}
			}
		}
	} else {
		log.Logger.Info("DynamoDB not configured, audit service disabled")
	}

	// Create error channel and HTTP close channel
	errChan := make(chan error, 1)
	httpCloseCh := make(chan struct{}, 1)

	// Start HTTP server
	log.Logger.Info("Starting HTTP server",
		zap.String("address", config.GlobalConfig.HTTPServer.Addr))
	go http_server.Start(ctx, errChan, httpCloseCh, services)
	log.Logger.Info("HTTP server started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Logger.Error("Server error", zap.Error(err))
	case sig := <-sigChan:
		log.Logger.Info("Received signal", zap.String("signal", sig.String()))
	}

	log.Logger.Info("Shutting down server")
	cancel()

	if kafkaConsumer != nil {
		log.Logger.Info("Stopping Kafka consumer")
		if err := kafkaConsumer.Stop(); err != nil {
			log.Logger.Error("Failed to stop Kafka consumer", zap.Error(err))
		}
	}

	if kafkaProducer != nil {
		log.Logger.Info("Closing Kafka producer")
		if err := kafkaProducer.Close(); err != nil {
			log.Logger.Error("Failed to close Kafka producer", zap.Error(err))
		}
	}

	shutdownTimeout := DefaultShutdownTimeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	select {
	case <-httpCloseCh:
		log.Logger.Info("HTTP server shutdown completed")
	case <-shutdownCtx.Done():
		log.Logger.Warn("HTTP server shutdown timed out",
			zap.Duration("timeout", DefaultShutdownTimeout))
	}

	log.Logger.Info("Server gracefully stopped")
}
