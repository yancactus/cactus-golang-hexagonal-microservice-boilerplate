package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cactus-golang-hexagonal-microservice-boilerplate/api/dto"
	"cactus-golang-hexagonal-microservice-boilerplate/api/http/handle"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
)

// User Handlers

// CreateUser creates a new user
func CreateUser(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handle.Error(c, err)
		return
	}

	user, err := services.UserService.Create(c.Request.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		handle.Error(c, err)
		return
	}

	handle.Success(c, toUserResp(user))
}

// GetUser retrieves a user by ID
func GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := services.UserService.Get(c.Request.Context(), id)
	if err != nil {
		handle.Error(c, err)
		return
	}
	if user == nil {
		handle.Error(c, model.ErrUserNotFound)
		return
	}

	handle.Success(c, toUserResp(user))
}

// UpdateUser updates an existing user
func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handle.Error(c, err)
		return
	}

	user, err := services.UserService.Update(c.Request.Context(), id, req.Name)
	if err != nil {
		handle.Error(c, err)
		return
	}

	handle.Success(c, toUserResp(user))
}

// DeleteUser deletes a user
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := services.UserService.Delete(c.Request.Context(), id); err != nil {
		handle.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// ListUsers lists all users with pagination
func ListUsers(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users, total, err := services.UserService.List(c.Request.Context(), offset, limit)
	if err != nil {
		handle.Error(c, err)
		return
	}

	resp := make([]*dto.UserResp, len(users))
	for i, u := range users {
		resp[i] = toUserResp(u)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  resp,
		"total": total,
	})
}

// GetUserOrders retrieves orders for a user
func GetUserOrders(c *gin.Context) {
	id := c.Param("id")

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orders, total, err := services.OrderService.GetByUserID(c.Request.Context(), id, offset, limit)
	if err != nil {
		handle.Error(c, err)
		return
	}

	resp := make([]*dto.OrderResp, len(orders))
	for i, o := range orders {
		resp[i] = toOrderResp(o)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  resp,
		"total": total,
	})
}

// Product Handlers

// CreateProduct creates a new product
func CreateProduct(c *gin.Context) {
	var req dto.CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handle.Error(c, err)
		return
	}

	product, err := services.ProductService.Create(c.Request.Context(), req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		handle.Error(c, err)
		return
	}

	handle.Success(c, toProductResp(product))
}

// GetProduct retrieves a product by ID
func GetProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := services.ProductService.Get(c.Request.Context(), id)
	if err != nil {
		handle.Error(c, err)
		return
	}
	if product == nil {
		handle.Error(c, model.ErrProductNotFound)
		return
	}

	handle.Success(c, toProductResp(product))
}

// UpdateProduct updates an existing product
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handle.Error(c, err)
		return
	}

	product, err := services.ProductService.Update(c.Request.Context(), id, req.Name, req.Description, req.Price)
	if err != nil {
		handle.Error(c, err)
		return
	}

	handle.Success(c, toProductResp(product))
}

// DeleteProduct deletes a product
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	if err := services.ProductService.Delete(c.Request.Context(), id); err != nil {
		handle.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

// ListProducts lists all products with pagination
func ListProducts(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, total, err := services.ProductService.List(c.Request.Context(), offset, limit)
	if err != nil {
		handle.Error(c, err)
		return
	}

	resp := make([]*dto.ProductResp, len(products))
	for i, p := range products {
		resp[i] = toProductResp(p)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  resp,
		"total": total,
	})
}

// UpdateProductStock updates product stock
func UpdateProductStock(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handle.Error(c, err)
		return
	}

	if err := services.ProductService.UpdateStock(c.Request.Context(), id, req.Quantity); err != nil {
		handle.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock updated"})
}

// Order Handlers

// CreateOrder creates a new order
func CreateOrder(c *gin.Context) {
	var req dto.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handle.Error(c, err)
		return
	}

	items := make([]model.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = model.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	order, err := services.OrderService.Create(c.Request.Context(), req.UserID, items)
	if err != nil {
		handle.Error(c, err)
		return
	}

	handle.Success(c, toOrderResp(order))
}

// GetOrder retrieves an order by ID
func GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := services.OrderService.Get(c.Request.Context(), id)
	if err != nil {
		handle.Error(c, err)
		return
	}
	if order == nil {
		handle.Error(c, model.ErrOrderNotFound)
		return
	}

	handle.Success(c, toOrderResp(order))
}

// ListOrders lists all orders with pagination
func ListOrders(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orders, total, err := services.OrderService.List(c.Request.Context(), offset, limit)
	if err != nil {
		handle.Error(c, err)
		return
	}

	resp := make([]*dto.OrderResp, len(orders))
	for i, o := range orders {
		resp[i] = toOrderResp(o)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  resp,
		"total": total,
	})
}

// UpdateOrderStatus updates order status
func UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateOrderStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handle.Error(c, err)
		return
	}

	if err := services.OrderService.UpdateStatus(c.Request.Context(), id, model.OrderStatus(req.Status)); err != nil {
		handle.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// CancelOrder cancels an order
func CancelOrder(c *gin.Context) {
	id := c.Param("id")

	if err := services.OrderService.Cancel(c.Request.Context(), id); err != nil {
		handle.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order canceled"})
}

// Helper functions to convert models to DTOs

func toUserResp(u *model.User) *dto.UserResp {
	return &dto.UserResp{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toProductResp(p *model.Product) *dto.ProductResp {
	return &dto.ProductResp{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toOrderResp(o *model.Order) *dto.OrderResp {
	items := make([]dto.OrderItemResp, len(o.Items))
	for i, item := range o.Items {
		items[i] = dto.OrderItemResp{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &dto.OrderResp{
		ID:        o.ID,
		UserID:    o.UserID,
		Items:     items,
		Total:     o.Total,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

// Audit Handlers

// GetAuditLog retrieves an audit log by ID
func GetAuditLog(c *gin.Context) {
	id := c.Param("id")

	if services.AuditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Audit service not available. DynamoDB may not be configured."})
		return
	}

	audit, err := services.AuditService.GetByID(c.Request.Context(), id)
	if err != nil {
		handle.Error(c, err)
		return
	}
	if audit == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
		return
	}

	handle.Success(c, toAuditResp(audit))
}

// ListAuditLogs lists audit logs by entity type with pagination
func ListAuditLogs(c *gin.Context) {
	if services.AuditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Audit service not available. DynamoDB may not be configured."})
		return
	}

	entityType := c.DefaultQuery("entity_type", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	lastKey := c.DefaultQuery("last_key", "")

	if entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_type query parameter is required"})
		return
	}

	audits, nextKey, err := services.AuditService.GetByEntityType(c.Request.Context(), entityType, limit, lastKey)
	if err != nil {
		handle.Error(c, err)
		return
	}

	resp := make([]*dto.AuditLogResp, len(audits))
	for i, a := range audits {
		resp[i] = toAuditResp(a)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     resp,
		"next_key": nextKey,
	})
}

// GetEntityAuditLogs retrieves audit logs for a specific entity
func GetEntityAuditLogs(c *gin.Context) {
	if services.AuditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Audit service not available. DynamoDB may not be configured."})
		return
	}

	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")

	audits, err := services.AuditService.GetByEntityID(c.Request.Context(), entityType, entityID)
	if err != nil {
		handle.Error(c, err)
		return
	}

	resp := make([]*dto.AuditLogResp, len(audits))
	for i, a := range audits {
		resp[i] = toAuditResp(a)
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func toAuditResp(a *model.AuditLog) *dto.AuditLogResp {
	return &dto.AuditLogResp{
		ID:         a.ID,
		EntityType: a.EntityType,
		EntityID:   a.EntityID,
		Action:     a.Action,
		Payload:    a.Payload,
		Timestamp:  a.Timestamp,
		UserID:     a.UserID,
	}
}
