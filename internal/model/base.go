package model

import "time"

// Entity provides common fields for database entities.
// Inspired by gorm.Model, but simplified for our use case.
// See: https://gorm.io/docs/models.html#gorm-Model
type Entity struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
}
