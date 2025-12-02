package models

import "time"

type ChannelType string
type StatusType string

const (
	ChannelEmail    ChannelType = "email"
	ChannelSMS      ChannelType = "sms"
	ChannelWhatsApp ChannelType = "whatsapp"

	StatusQueued     StatusType = "queued"
	StatusProcessing StatusType = "processing"
	StatusSent       StatusType = "sent"
	StatusFailed     StatusType = "failed"
	StatusRetrying   StatusType = "retrying"
	StatusDead       StatusType = "dead"
)

type Notification struct {
	ID            uint        `json:"id" gorm:"primaryKey"`
	UserID        uint        `json:"userId" gorm:"not null;index"`
	Channel       ChannelType `json:"channel" gorm:"type:channel_type;not null"`
	To            string      `json:"to" gorm:"type:varchar(255);not null"`
	Subject       *string     `json:"subject,omitempty" gorm:"type:varchar(255)"`
	Body          string      `json:"body" gorm:"type:text"`
	Status        StatusType  `json:"status" gorm:"type:status_type"`
	Provider      string      `json:"provider,omitempty" gorm:"type:varchar(50)"`
	Error         string      `json:"error,omitempty" gorm:"type:text"`
	Retries       int         `json:"retries" gorm:"default:0"`
	MaxRetries    int         `json:"maxRetries" gorm:"default:3"`
	NextAttemptAt *time.Time  `json:"nextAttempt,omitempty"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}
