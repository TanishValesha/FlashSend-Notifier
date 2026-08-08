package router

import (
	"net/http"

	apikey "github.com/TanishValesha/FlashSend-Notifier/internal/apiKey"
	"github.com/TanishValesha/FlashSend-Notifier/internal/auth"
	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	"github.com/TanishValesha/FlashSend-Notifier/internal/logger"
	notify "github.com/TanishValesha/FlashSend-Notifier/internal/notify"
	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Init(version, buildTime string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Ping godoc
	// @Summary      Ping
	// @Description  Simple health check endpoint.
	// @Tags         System
	// @Produce      json
	// @Success      200  {object}  object{message=string}
	// @Router       /ping [get]
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Health godoc
	// @Summary      Health Check
	// @Description  Checks DB and RabbitMQ connectivity. Used by Kubernetes liveness/readiness probes.
	// @Tags         System
	// @Produce      json
	// @Success      200  {object}  object{status=string,buildTime=string,db=string,rabbitmq=string,version=string}
	// @Failure      503  {object}  object{status=string,buildTime=string,db=string,rabbitmq=string,version=string}
	// @Router       /health [get]
	router.GET("/health", func(c *gin.Context) {
		dbStatus := "ok"
		mqStatus := "ok"

		if err := db.DB.Raw("SELECT 1").Error; err != nil {
			dbStatus = "unhealthy"
		}

		if !rabbitmq.IsConnected() {
			mqStatus = "unhealthy"
		}

		status := http.StatusOK
		if dbStatus != "ok" || mqStatus != "ok" {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":    map[bool]string{true: "healthy", false: "unhealthy"}[status == http.StatusOK],
			"buildTime": buildTime,
			"db":        dbStatus,
			"rabbitmq":  mqStatus,
			"version":   version,
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
		// GetUser godoc
		// @Summary      Get Current User
		// @Description  Returns the authenticated user's ID and email.
		// @Tags         User
		// @Produce      json
		// @Security     BearerAuth
		// @Success      200  {object}  object{user_id=float64,email=string}
		// @Router       /api/get-user [get]
		protected.GET("/get-user", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id": c.GetFloat64("user_id"),
				"email":   c.GetString("email"),
			})
		})

		// GetLogs godoc
		// @Summary      Get Logs
		// @Description  Returns recent notification logs for the authenticated user.
		// @Tags         Logs
		// @Produce      json
		// @Security     BearerAuth
		// @Success      200  {object}  object{logs=[]models.Notification}
		// @Router       /api/logs [get]
		protected.GET("/logs", logger.GetLogsHandler)

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
			notifyGroup.POST("/sms", notify.SMSNotifyHandler)
			notifyGroup.POST("/send", notify.UnifiedNotifyHandler)
			notifyGroup.POST("/schedule", notify.ScheduledNotificationHandler)
		}
	}

	return router
}
