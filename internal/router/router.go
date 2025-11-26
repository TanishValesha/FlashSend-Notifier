package router

import (
	"net/http"

	"github.com/TanishValesha/FlashSend-Notifier/internal/auth"
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

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
	}

	return router
}
