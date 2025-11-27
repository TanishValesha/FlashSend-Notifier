package notify

import (
	"fmt"
	"net/smtp"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func SendEmail(to, subject, body string) error {
	from := config.Cfg.SMTPEmail
	password := config.Cfg.SMTPAppPassword
	smtpHost := config.Cfg.SMTPHost
	smtpPort := config.Cfg.SMTPPort

	message := "From: " + from + "\n" + "To: " + to + "\n" + "Subject: " + subject + "\n\n" + body

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("email send error: %v", err)
	}

	return nil
}
