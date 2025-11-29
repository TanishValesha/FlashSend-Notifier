package unified

import (
	"errors"

	"github.com/TanishValesha/FlashSend-Notifier/internal/models"
	email "github.com/TanishValesha/FlashSend-Notifier/internal/notify/email"
	"github.com/TanishValesha/FlashSend-Notifier/internal/notify/sms"
)

func SendUnifiedNotification(req models.UnifiedRequest) error {
	if req.Channel != "sms" && req.Channel != "email" {
		return errors.New("invalid channel selection (sms OR email)")
	}

	if req.To == "" {
		return errors.New("provide a receiver")
	}

	if req.Body == nil {
		return errors.New("body is required")
	}

	if req.Channel == "email" && req.Subject == nil {
		return errors.New("subject is required for email")
	}

	if req.Channel == "sms" {
		return sms.SendSMS(req.To, *req.Body)
	}

	if req.Channel == "email" {
		return email.SendEmail(req.To, *req.Subject, *req.Body)
	}

	return nil
}
