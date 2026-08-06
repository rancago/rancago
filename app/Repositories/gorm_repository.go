package Repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/rancago/framework/app/Contracts"
)

// GormRepository is a generic GORM-backed implementation of Contracts.Repository[T].
// T must be a GORM model struct (e.g. Models.User, Models.Document).
//
// Example:
//
//	userRepo := Repositories.NewGormRepository[Models.User](db)
//	user, err := userRepo.FindByID(ctx, 1)
type GormRepository[T any] struct {
	db *gorm.DB
}

// NewGormRepository creates a GormRepository backed by the given *gorm.DB.
func NewGormRepository[T any](db *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{db: db}
}

// FindByID retrieves a single record by primary key.
func (r *GormRepository[T]) FindByID(ctx context.Context, id interface{}) (*T, error) {
	var entity T
	result := r.db.WithContext(ctx).First(&entity, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GormRepository.FindByID: %w", result.Error)
	}
	return &entity, nil
}

// FindAll retrieves all records.
func (r *GormRepository[T]) FindAll(ctx context.Context) ([]*T, error) {
	var entities []*T
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("GormRepository.FindAll: %w", err)
	}
	return entities, nil
}

// FindPaginated retrieves a page of records.
func (r *GormRepository[T]) FindPaginated(ctx context.Context, page, perPage int) ([]*T, Contracts.PaginationMeta, error) {
	if perPage <= 0 {
		perPage = 25
	}
	if page < 1 {
		page = 1
	}

	var total int64
	var zero T
	if err := r.db.WithContext(ctx).Model(&zero).Count(&total).Error; err != nil {
		return nil, Contracts.PaginationMeta{}, fmt.Errorf("GormRepository.FindPaginated count: %w", err)
	}

	offset := (page - 1) * perPage
	var entities []*T
	if err := r.db.WithContext(ctx).Offset(offset).Limit(perPage).Find(&entities).Error; err != nil {
		return nil, Contracts.PaginationMeta{}, fmt.Errorf("GormRepository.FindPaginated: %w", err)
	}

	totalPages := (total + int64(perPage) - 1) / int64(perPage)
	meta := Contracts.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(offset+perPage) < total,
		HasPrev:    offset > 0,
	}
	return entities, meta, nil
}

// Create inserts a new record and returns the saved entity.
func (r *GormRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, fmt.Errorf("GormRepository.Create: %w", err)
	}
	return entity, nil
}

// Update saves all non-zero fields of entity (uses GORM Save).
func (r *GormRepository[T]) Update(ctx context.Context, entity *T) (*T, error) {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, fmt.Errorf("GormRepository.Update: %w", err)
	}
	return entity, nil
}

// Delete removes a record by primary key.
func (r *GormRepository[T]) Delete(ctx context.Context, id interface{}) error {
	var entity T
	if err := r.db.WithContext(ctx).Delete(&entity, id).Error; err != nil {
		return fmt.Errorf("GormRepository.Delete: %w", err)
	}
	return nil
}

// FindBy retrieves records matching the given conditions map.
func (r *GormRepository[T]) FindBy(ctx context.Context, conditions map[string]interface{}) ([]*T, error) {
	var entities []*T
	if err := r.db.WithContext(ctx).Where(conditions).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("GormRepository.FindBy: %w", err)
	}
	return entities, nil
}

// FirstOrCreate finds the first record matching conditions or creates one with defaults.
func (r *GormRepository[T]) FirstOrCreate(
	ctx context.Context,
	conditions, defaults map[string]interface{},
) (*T, bool, error) {
	var entity T

	// Build attrs = conditions merged with defaults
	attrs := make(map[string]interface{}, len(defaults))
	for k, v := range defaults {
		attrs[k] = v
	}

	result := r.db.WithContext(ctx).Where(conditions).Attrs(attrs).FirstOrCreate(&entity)
	if result.Error != nil {
		return nil, false, fmt.Errorf("GormRepository.FirstOrCreate: %w", result.Error)
	}
	created := result.RowsAffected > 0
	return &entity, created, nil
}

// DB returns the underlying *gorm.DB for advanced queries.
func (r *GormRepository[T]) DB() *gorm.DB {
	return r.db
}

// Scoped returns a new GormRepository using a scoped *gorm.DB (e.g. with Where clause).
func (r *GormRepository[T]) Scoped(scope func(*gorm.DB) *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{db: scope(r.db)}
}
