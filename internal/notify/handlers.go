package notify

import (
	"fmt"
	"net/http"
	"time"

	utils "github.com/TanishValesha/FlashSend-Notifier/internal"
	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	"github.com/TanishValesha/FlashSend-Notifier/internal/logger"
	models "github.com/TanishValesha/FlashSend-Notifier/internal/models"
	email "github.com/TanishValesha/FlashSend-Notifier/internal/notify/email"
	sms "github.com/TanishValesha/FlashSend-Notifier/internal/notify/sms"
	"github.com/TanishValesha/FlashSend-Notifier/internal/notify/unified"
	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
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

	entry = models.Notification{
		UserID:   user_id,
		Channel:  "email",
		To:       req.To,
		Subject:  &req.Subject,
		Body:     req.Body,
		Status:   models.StatusQueued,
		Provider: "smtp",
	}

	logger.LogNotification(&entry)

	msg := rabbitmq.QueueMessage{
		NotificationID:      entry.ID,
		NotificationChannel: rabbitmq.ChannelEmail,
		To:                  req.To,
		Subject:             req.Subject,
		Body:                req.Body,
	}

	if err := rabbitmq.PublishMessageToQueue(msg); err != nil {
		entry.Status = models.StatusFailed
		entry.Error = err.Error()
		db.DB.Save(&entry)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue publish failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email queued and will be sent shortly",
		"id":      entry.ID,
	})
}

func SMSNotifyHandler(c *gin.Context) {
	var req sms.SMSRequest
	var entry models.Notification
	user_id := uint(c.GetFloat64("user_id"))
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	entry = models.Notification{
		UserID:   user_id,
		Channel:  "sms",
		To:       req.To,
		Body:     req.Body,
		Status:   models.StatusQueued,
		Provider: "twilio",
	}

	logger.LogNotification(&entry)

	msg := rabbitmq.QueueMessage{
		NotificationID:      entry.ID,
		NotificationChannel: rabbitmq.ChannelSMS,
		To:                  req.To,
		Body:                req.Body,
	}

	if err := rabbitmq.PublishMessageToQueue(msg); err != nil {
		entry.Status = models.StatusFailed
		entry.Error = err.Error()
		db.DB.Save(&entry)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue publish failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SMS queued and will be sent shortly",
		"id":      entry.ID,
	})
}

func ScheduledNotificationHandler(c *gin.Context) {
	var req struct {
		Channel     string  `json:"channel"`
		To          string  `json:"to"`
		Subject     *string `json:"subject,omitempty"`
		Body        string  `json:"body"`
		ScheduledAt *string `json:"scheduled_at,omitempty"`
	}
	user_id := uint(c.GetFloat64("user_id"))
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	parsedTime, err := utils.ParseDateTime(*req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if parsedTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Scheduled time must be in the future"})
		return
	}

	entry := models.Notification{
		UserID:      user_id,
		Channel:     models.ChannelType(req.Channel),
		To:          req.To,
		Subject:     req.Subject,
		Body:        req.Body,
		IsScheduled: true,
		ScheduledAt: parsedTime,
		Status:      models.StatusScheduled,
		MaxRetries:  3,
	}

	if err := db.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB insert failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Scheduled notification created",
		"id":      entry.ID,
	})
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
		logger.LogNotification(&entry)
		return
	}

	logger.LogNotification(&entry)

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
}
