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
	Attempts            int         `json:"attempts" gorm:"default:1"`
}
