// Package model contains the persisted domain types.
package model

import "time"

// Item is the single resource this template exposes. Replace it with your own
// domain type, or keep it as the worked example.
type Item struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:120;not null"`
	Description string    `json:"description" gorm:"size:2000"`
	Status      string    `json:"status" gorm:"size:20;not null;default:active;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Statuses an item may be in. Kept as constants so the service and the API
// validation cannot drift.
const (
	StatusActive   = "active"
	StatusArchived = "archived"
)

// ValidStatus reports whether the value is one the API accepts.
func ValidStatus(value string) bool {
	return value == StatusActive || value == StatusArchived
}
