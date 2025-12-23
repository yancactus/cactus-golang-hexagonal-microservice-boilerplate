package http

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	httpMiddleware "cactus-golang-hexagonal-microservice-boilerplate/api/http/middleware"
	"cactus-golang-hexagonal-microservice-boilerplate/api/http/validator/custom"
	metricsMiddleware "cactus-golang-hexagonal-microservice-boilerplate/api/middleware"
	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// Service instances for API handlers
var services *service.Services

// RegisterServices registers service instances for API handlers
func RegisterServices(s *service.Services) {
	services = s
}

// NewServerRoute creates and configures the HTTP server routes
func NewServerRoute() *gin.Engine {
	if config.GlobalConfig.Env.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		custom.RegisterValidators(v)
	}

	// Apply middleware
	applyMiddleware(router)

	// Health check
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Debug tools
	if config.GlobalConfig.HTTPServer.Pprof {
		httpMiddleware.RegisterPprof(router)
	}

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register API routes
	registerAPIRoutes(router)

	return router
}

// applyMiddleware applies all middleware to the router
func applyMiddleware(router *gin.Engine) {
	router.Use(gin.Recovery())
	router.Use(httpMiddleware.RequestID())
	router.Use(httpMiddleware.Cors())
	router.Use(httpMiddleware.RequestLogger())
	router.Use(httpMiddleware.Translations())
	router.Use(httpMiddleware.EnhancedErrorHandlerMiddleware())

	// Add metrics middleware
	router.Use(func(c *gin.Context) {
		handlerName := c.FullPath()
		if handlerName == "" {
			handlerName = "unknown"
		}

		start := time.Now()
		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		metricsMiddleware.RecordHTTPMetrics(handlerName, c.Request.Method, statusCode, duration)
	})
}

// registerAPIRoutes registers all API routes
func registerAPIRoutes(router *gin.Engine) {
	api := router.Group("/api")

	// User API
	users := api.Group("/users")
	users.POST("", CreateUser)
	users.GET("", ListUsers)
	users.GET("/:id", GetUser)
	users.PUT("/:id", UpdateUser)
	users.DELETE("/:id", DeleteUser)
	users.GET("/:id/orders", GetUserOrders)

	// Product API
	products := api.Group("/products")
	products.POST("", CreateProduct)
	products.GET("", ListProducts)
	products.GET("/:id", GetProduct)
	products.PUT("/:id", UpdateProduct)
	products.DELETE("/:id", DeleteProduct)
	products.PATCH("/:id/stock", UpdateProductStock)

	// Order API
	orders := api.Group("/orders")
	orders.POST("", CreateOrder)
	orders.GET("", ListOrders)
	orders.GET("/:id", GetOrder)
	orders.PATCH("/:id/status", UpdateOrderStatus)
	orders.POST("/:id/cancel", CancelOrder)

	// Audit API
	audits := api.Group("/audits")
	audits.GET("", ListAuditLogs)
	audits.GET("/log/:id", GetAuditLog)
	audits.GET("/entity/:entity_type/:entity_id", GetEntityAuditLogs)
}
