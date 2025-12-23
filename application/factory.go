package application

import (
	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// Factory provides access to application services
type Factory struct {
	services *service.Services
}

// NewFactory creates a new application factory
func NewFactory(services *service.Services) *Factory {
	return &Factory{
		services: services,
	}
}

// UserService returns the user service
func (f *Factory) UserService() service.IUserService {
	return f.services.UserService
}

// ProductService returns the product service
func (f *Factory) ProductService() service.IProductService {
	return f.services.ProductService
}

// OrderService returns the order service
func (f *Factory) OrderService() service.IOrderService {
	return f.services.OrderService
}
