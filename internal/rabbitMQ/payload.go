package rabbitmq

type ChannelType string

const (
	ChannelEmail    ChannelType = "email"
	ChannelSMS      ChannelType = "sms"
	ChannelWhatsApp ChannelType = "whatsapp"
)

type QueueMessage struct {
	NotificationID      uint        `json:"notification_id" gorm:"primaryKey"`
	NotificationChannel ChannelType `json:"notification_channel"`
	To                  string      `json:"to"`
	Subject             string      `json:"subject,omitempty"`
	Body                string      `json:"body,omitempty"`
	Retries             int         `json:"retries" gorm:"default:0"`
}
