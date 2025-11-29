package logger

import (
	"net/http"

	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	"github.com/TanishValesha/FlashSend-Notifier/internal/models"
	"github.com/gin-gonic/gin"
)

func LogNotification(entry models.Notification) error {
	return db.DB.Create(&entry).Error
}

func GetLogsHandler(c *gin.Context) {
	userID := uint(c.GetFloat64("user_id"))

	var logs []models.Notification

	db.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
	})
}
