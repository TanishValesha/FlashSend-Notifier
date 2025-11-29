package models

type UnifiedRequest struct {
	Channel string  `json:"channel" binding:"required,oneof=sms email"`
	To      string  `json:"to" binding:"required"`
	Subject *string `json:"subject"`
	Body    *string `json:"body"`
}
