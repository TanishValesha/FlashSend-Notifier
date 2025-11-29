package models

import "time"

type APIKey struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"index;not null"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null"`
	Active    bool      `json:"active" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt"`
}
