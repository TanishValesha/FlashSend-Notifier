package router

import (
	"net/http"

	apikey "github.com/TanishValesha/FlashSend-Notifier/internal/apiKey"
	"github.com/TanishValesha/FlashSend-Notifier/internal/auth"
	"github.com/TanishValesha/FlashSend-Notifier/internal/notify"
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	router := gin.Default()

	apiGroup := router.Group("/api")

	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register", auth.RegisterHandler)
		authGroup.POST("/login", auth.LoginHandler)
	}

	protected := apiGroup.Group("/")
	protected.Use(auth.JWTMiddleware())
	{
		protected.GET("/get-user", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id": c.GetFloat64("user_id"),
				"email":   c.GetString("email"),
			})
		})

		keysGroup := protected.Group("/keys")
		{
			keysGroup.POST("/", apikey.CreateAPIKeyHandler)
			keysGroup.GET("/", apikey.ListAllAPIKeys)
			keysGroup.DELETE("/:id", apikey.DeleteAPIKeyHandler)
			keysGroup.PATCH("/toggle/:id", apikey.ToggleAPIKey)
		}

		notifyGroup := protected.Group("/notify")
		notifyGroup.Use(notify.APIKeyMiddleware())
		{
			notifyGroup.POST("/email", notify.EmailNotifyHandler)
		}
	}

	return router
}
