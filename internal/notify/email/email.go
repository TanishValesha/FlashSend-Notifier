package notify

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func SendEmail(to, subject, body string) error {
	from := config.Cfg.SMTPEmail
	password := config.Cfg.SMTPAppPassword
	smtpHost := config.Cfg.SMTPHost // e.g. smtp.gmail.com
	smtpPort := config.Cfg.SMTPPort // usually 587

	// Build email
	message := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	// Force IPv4-only dialer
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		DualStack: false,
	}

	// Connect via IPv4
	conn, err := dialer.Dial("tcp4", smtpHost+":"+smtpPort)
	if err != nil {
		return fmt.Errorf("IPv4 connection failed: %w", err)
	}

	// Create SMTP client on the raw connection
	c, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("SMTP client failed: %w", err)
	}

	// Gmail requires EHLO before STARTTLS
	if err = c.Hello("localhost"); err != nil {
		return fmt.Errorf("EHLO failed: %w", err)
	}

	// STARTTLS is mandatory for Gmail on port 587
	tlsConfig := &tls.Config{ServerName: smtpHost}
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	} else {
		return fmt.Errorf("SMTP server does not support STARTTLS")
	}

	// Authenticate after TLS
	auth := smtp.PlainAuth("", from, password, smtpHost)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	// Set from + to
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	// Write message
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	c.Quit()
	return nil
}
