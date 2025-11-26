package models

import "time"

type APIKey struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index;not null"`
	Key       string `gorm:"uniqueIndex;not null"`
	Active    bool   `gorm:"default:true"`
	CreatedAt time.Time
}
