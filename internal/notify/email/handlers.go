package notify

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func EmailNotifyHandler(c *gin.Context) {
	var req EmailRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		fmt.Print(err)
		return
	}

	err := SendEmail(req.To, req.Subject, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Print(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent"})
}
