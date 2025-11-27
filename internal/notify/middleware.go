package notify

import (
	"net/http"

	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	"github.com/TanishValesha/FlashSend-Notifier/internal/models"
	"github.com/gin-gonic/gin"
)

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API key"})
			c.Abort()
			return
		}

		var key models.APIKey
		if err := db.DB.Where("key = ?", apiKey).First(&key).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		if !key.Active {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Inactive API key"})
			c.Abort()
			return
		}

		c.Set("api_user_id", key.UserID)

		c.Next()

	}
}
