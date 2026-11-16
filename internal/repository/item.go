// Package repository is the only layer that knows about the database.
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"go-service-template/internal/model"
)

// ErrNotFound is returned by every lookup miss, so the service layer does not
// need to import gorm.
var ErrNotFound = errors.New("item not found")

// ItemStore is the persistence contract. Tests can replace it with a fake.
type ItemStore interface {
	Create(ctx context.Context, item *model.Item) error
	Get(ctx context.Context, id uint) (model.Item, error)
	List(ctx context.Context, limit, offset int) ([]model.Item, int64, error)
	Update(ctx context.Context, item *model.Item) error
	Delete(ctx context.Context, id uint) error
}

// GormItemStore is the SQLite implementation.
type GormItemStore struct {
	db *gorm.DB
}

// NewItemStore wires the store to a database handle.
func NewItemStore(db *gorm.DB) *GormItemStore {
	return &GormItemStore{db: db}
}

// Create inserts one row and fills in the generated ID and timestamps.
func (s *GormItemStore) Create(ctx context.Context, item *model.Item) error {
	return s.db.WithContext(ctx).Create(item).Error
}

// Get returns one row or ErrNotFound.
func (s *GormItemStore) Get(ctx context.Context, id uint) (model.Item, error) {
	var item model.Item
	err := s.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Item{}, ErrNotFound
	}
	return item, err
}

// List returns a page of rows, newest first, plus the total count.
func (s *GormItemStore) List(ctx context.Context, limit, offset int) ([]model.Item, int64, error) {
	var items []model.Item
	var total int64
	query := s.db.WithContext(ctx).Model(&model.Item{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}

// Update writes the mutable columns. It deliberately avoids gorm.Save, which
// upserts: an update to a missing row must report a miss, not create a ghost.
func (s *GormItemStore) Update(ctx context.Context, item *model.Item) error {
	result := s.db.WithContext(ctx).
		Model(&model.Item{}).
		Where("id = ?", item.ID).
		Select("name", "description", "status", "updated_at").
		Updates(item)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes one row and reports a miss as ErrNotFound.
func (s *GormItemStore) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&model.Item{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
