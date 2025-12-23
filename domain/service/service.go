package service

import (
	"cactus-golang-hexagonal-microservice-boilerplate/domain/event"
)

// Services contains all service instances
type Services struct {
	UserService    IUserService
	ProductService IProductService
	OrderService   IOrderService
	AuditService   IAuditService
	EventBus       event.EventBus
}

// NewServices creates a services collection
func NewServices(userService IUserService, productService IProductService, orderService IOrderService, eventBus event.EventBus) *Services {
	return &Services{
		UserService:    userService,
		ProductService: productService,
		OrderService:   orderService,
		EventBus:       eventBus,
	}
}
