package postgre

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
)

// UserRepository implements IUserRepo using PostgreSQL
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) repo.IUserRepo {
	return &UserRepository{db: db}
}

// userEntity represents the database entity
type userEntity struct {
	ID        string     `gorm:"primaryKey;type:uuid"`
	Email     string     `gorm:"uniqueIndex;not null"`
	Name      string     `gorm:"not null"`
	Password  string     `gorm:"not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

func (userEntity) TableName() string {
	return "users"
}

// toModel converts entity to domain model
func (e *userEntity) toModel() *model.User {
	return &model.User{
		ID:        e.ID,
		Email:     e.Email,
		Name:      e.Name,
		Password:  e.Password,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: e.DeletedAt,
	}
}

// toEntity converts domain model to entity
func toUserEntity(u *model.User) *userEntity {
	return &userEntity{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeletedAt: u.DeletedAt,
	}
}

func (r *UserRepository) getDB(ctx context.Context, tx repo.Transaction) *gorm.DB {
	if tx != nil {
		if gormTx, ok := tx.GetTx().(*gorm.DB); ok {
			return gormTx.WithContext(ctx)
		}
	}
	return r.db.WithContext(ctx)
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, tx repo.Transaction, user *model.User) (*model.User, error) {
	entity := toUserEntity(user)
	db := r.getDB(ctx, tx)

	if err := db.Create(entity).Error; err != nil {
		return nil, err
	}

	return entity.toModel(), nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, tx repo.Transaction, user *model.User) error {
	entity := toUserEntity(user)
	db := r.getDB(ctx, tx)

	return db.Save(entity).Error
}

// Delete soft deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, tx repo.Transaction, id string) error {
	db := r.getDB(ctx, tx)

	now := time.Now()
	return db.Model(&userEntity{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, tx repo.Transaction, id string) (*model.User, error) {
	var entity userEntity
	db := r.getDB(ctx, tx)

	err := db.Where("id = ? AND deleted_at IS NULL", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return entity.toModel(), nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, tx repo.Transaction, email string) (*model.User, error) {
	var entity userEntity
	db := r.getDB(ctx, tx)

	err := db.Where("email = ? AND deleted_at IS NULL", email).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return entity.toModel(), nil
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, tx repo.Transaction, offset, limit int) ([]*model.User, int64, error) {
	var entities []userEntity
	var total int64
	db := r.getDB(ctx, tx)

	// Get total count
	if err := db.Model(&userEntity{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := db.Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*model.User, len(entities))
	for i, e := range entities {
		users[i] = e.toModel()
	}

	return users, total, nil
}
