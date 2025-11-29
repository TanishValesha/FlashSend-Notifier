package notify

import (
	"fmt"
	"net/http"

	"github.com/TanishValesha/FlashSend-Notifier/internal/logger"
	models "github.com/TanishValesha/FlashSend-Notifier/internal/models"
	email "github.com/TanishValesha/FlashSend-Notifier/internal/notify/email"
	sms "github.com/TanishValesha/FlashSend-Notifier/internal/notify/sms"
	"github.com/TanishValesha/FlashSend-Notifier/internal/notify/unified"
	"github.com/gin-gonic/gin"
)

func EmailNotifyHandler(c *gin.Context) {
	var req email.EmailRequest
	var entry models.Notification
	user_id := uint(c.GetFloat64("user_id"))
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		fmt.Print(err)
		return
	}

	err := email.SendEmail(req.To, req.Subject, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Print(err)
		entry = models.Notification{
			UserID:  user_id,
			Channel: "email",
			To:      req.To,
			Subject: &req.Subject,
			Body:    req.Body,
			Status:  "failed",
			Error:   err.Error(),
		}
		logger.LogNotification(entry)
		return
	}

	entry = models.Notification{
		UserID:  user_id,
		Channel: "email",
		To:      req.To,
		Subject: &req.Subject,
		Body:    req.Body,
		Status:  "sent",
	}

	logger.LogNotification(entry)

	c.JSON(http.StatusOK, gin.H{"message": "Email sent"})
}

func SMSNotifyHandler(c *gin.Context) {
	var req sms.SMSRequest
	var entry models.Notification
	user_id := uint(c.GetFloat64("user_id"))
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	err := sms.SendSMS(req.To, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		entry = models.Notification{
			UserID:  user_id,
			Channel: "sms",
			To:      req.To,
			Body:    req.Body,
			Status:  "failed",
			Error:   err.Error(),
		}
		logger.LogNotification(entry)
		return
	}

	entry = models.Notification{
		UserID:  user_id,
		Channel: "sms",
		To:      req.To,
		Body:    req.Body,
		Status:  "sent",
	}

	logger.LogNotification(entry)

	c.JSON(http.StatusOK, gin.H{"message": "SMS sent"})
}

func UnifiedNotifyHandler(c *gin.Context) {
	var req models.UnifiedRequest
	var entry models.Notification
	user_id := uint(c.GetFloat64("user_id"))
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	switch req.Channel {
	case "sms":
		entry = models.Notification{
			UserID:  user_id,
			Channel: "sms",
			To:      req.To,
			Body:    *req.Body,
			Status:  "sent",
		}
	case "email":
		entry = models.Notification{
			UserID:  user_id,
			Channel: "email",
			To:      req.To,
			Subject: req.Subject,
			Body:    *req.Body,
			Status:  "sent",
		}
	}

	err := unified.SendUnifiedNotification(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		entry.Status = "failed"
		entry.Error = err.Error()
		logger.LogNotification(entry)
		return
	}

	logger.LogNotification(entry)

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
}
