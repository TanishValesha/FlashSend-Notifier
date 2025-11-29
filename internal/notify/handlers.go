package notify

import (
	"fmt"
	"net/http"

	models "github.com/TanishValesha/FlashSend-Notifier/internal/models"
	email "github.com/TanishValesha/FlashSend-Notifier/internal/notify/email"
	sms "github.com/TanishValesha/FlashSend-Notifier/internal/notify/sms"
	"github.com/TanishValesha/FlashSend-Notifier/internal/notify/unified"
	"github.com/gin-gonic/gin"
)

func EmailNotifyHandler(c *gin.Context) {
	var req email.EmailRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		fmt.Print(err)
		return
	}

	err := email.SendEmail(req.To, req.Subject, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Print(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent"})
}

func SMSNotifyHandler(c *gin.Context) {
	var req sms.SMSRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	err := sms.SendSMS(req.To, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SMS sent"})
}

func UnifiedNotifyHandler(c *gin.Context) {
	var req models.UnifiedRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	err := unified.SendUnifiedNotification(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
}
