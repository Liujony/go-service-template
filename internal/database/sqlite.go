// Package database owns the SQLite connection and the startup migration.
package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go-service-template/internal/model"
)

// Open connects to SQLite, applies connection pragmas and migrates the schema.
// WAL plus a busy timeout is what makes a handful of concurrent writers behave.
func Open(path string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on", path)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.AutoMigrate(&model.Item{}); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}
	return db, nil
}

// Seed inserts a few rows so a fresh clone has something to show on the
// dashboard. It is a no-op when the table already has data.
func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Item{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	rows := []model.Item{
		{Name: "Ship the v1 API", Description: "CRUD, validation and the JSON envelope are in place.", Status: model.StatusActive},
		{Name: "Add request IDs", Description: "Every response carries X-Request-ID; logs carry the same value.", Status: model.StatusActive},
		{Name: "Old experiment", Description: "Kept for the archive view.", Status: model.StatusArchived},
	}
	return db.Create(&rows).Error
}
