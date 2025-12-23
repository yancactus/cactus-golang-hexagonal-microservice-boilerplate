package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
)

const productsCollection = "products"

// ProductRepository implements IProductRepo using MongoDB
type ProductRepository struct {
	client *Client
}

// NewProductRepository creates a new product repository
func NewProductRepository(client *Client) repo.IProductRepo {
	return &ProductRepository{client: client}
}

// productDocument represents the MongoDB document
type productDocument struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Price       float64            `bson:"price"`
	Stock       int                `bson:"stock"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty"`
}

// toModel converts document to domain model
func (d *productDocument) toModel() *model.Product {
	return &model.Product{
		ID:          d.ID.Hex(),
		Name:        d.Name,
		Description: d.Description,
		Price:       d.Price,
		Stock:       d.Stock,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		DeletedAt:   d.DeletedAt,
	}
}

// toDocument converts domain model to document
func toProductDocument(p *model.Product) (*productDocument, error) {
	doc := &productDocument{
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		DeletedAt:   p.DeletedAt,
	}

	if p.ID != "" {
		oid, err := primitive.ObjectIDFromHex(p.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}
		doc.ID = oid
	}

	return doc, nil
}

func (r *ProductRepository) collection() *mongo.Collection {
	return r.client.GetCollection(productsCollection)
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *model.Product) (*model.Product, error) {
	doc, err := toProductDocument(product)
	if err != nil {
		return nil, err
	}

	doc.ID = primitive.NewObjectID()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	_, err = r.collection().InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert product: %w", err)
	}

	return doc.toModel(), nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	oid, err := primitive.ObjectIDFromHex(product.ID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	filter := bson.M{"_id": oid, "deleted_at": nil}
	update := bson.M{
		"$set": bson.M{
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"stock":       product.Stock,
			"updated_at":  time.Now(),
		},
	}

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	if result.MatchedCount == 0 {
		return model.ErrProductNotFound
	}

	return nil
}

// Delete soft deletes a product by ID
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	now := time.Now()
	filter := bson.M{"_id": oid, "deleted_at": nil}
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.MatchedCount == 0 {
		return model.ErrProductNotFound
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	filter := bson.M{"_id": oid, "deleted_at": nil}

	var doc productDocument
	err = r.collection().FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return doc.toModel(), nil
}

// GetByName retrieves a product by name
func (r *ProductRepository) GetByName(ctx context.Context, name string) (*model.Product, error) {
	filter := bson.M{"name": name, "deleted_at": nil}

	var doc productDocument
	err := r.collection().FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return doc.toModel(), nil
}

// List retrieves products with pagination
func (r *ProductRepository) List(ctx context.Context, offset, limit int) ([]*model.Product, int64, error) {
	filter := bson.M{"deleted_at": nil}

	// Get total count
	total, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Get paginated results
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []*model.Product
	for cursor.Next(ctx) {
		var doc productDocument
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		products = append(products, doc.toModel())
	}

	return products, total, nil
}

// UpdateStock updates product stock
func (r *ProductRepository) UpdateStock(ctx context.Context, id string, quantity int) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	filter := bson.M{"_id": oid, "deleted_at": nil}
	update := bson.M{
		"$inc": bson.M{"stock": quantity},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	if result.MatchedCount == 0 {
		return model.ErrProductNotFound
	}

	return nil
}
