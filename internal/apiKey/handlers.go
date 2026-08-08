package apikey

import (
	"net/http"

	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	"github.com/TanishValesha/FlashSend-Notifier/internal/models"
	"github.com/gin-gonic/gin"
)

// CreateAPIKeyHandler generates a new API key for the authenticated user.
// @Summary      Create API Key
// @Description  Generates a new API key for the authenticated user.
// @Tags         API Keys
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{message=string,key=string}
// @Failure      500  {object}  object{error=string}
// @Router       /api/keys/ [post]
func CreateAPIKeyHandler(c *gin.Context) {
	user_id := uint(c.GetFloat64("user_id"))

	apiKey, err := GenerateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not generate API Key",
		})
		return
	}

	keyModel := models.APIKey{
		UserID: user_id,
		Key:    apiKey,
	}

	if err := db.DB.Create(&keyModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key generated",
		"key":     keyModel.Key,
	})
}

// ListAllAPIKeys lists all API keys for the authenticated user.
// @Summary      List API Keys
// @Description  Returns all API keys belonging to the authenticated user.
// @Tags         API Keys
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{keys=[]models.APIKey}
// @Failure      500  {object}  object{error=string}
// @Router       /api/keys/ [get]
func ListAllAPIKeys(c *gin.Context) {
	user_id := uint(c.GetFloat64("user_id"))

	var keys []models.APIKey

	if err := db.DB.Where("user_id = ?", user_id).Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

// DeleteAPIKeyHandler deletes an API key by ID.
// @Summary      Delete API Key
// @Description  Deletes a specific API key belonging to the authenticated user.
// @Tags         API Keys
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "API Key ID"
// @Success      200  {object}  object{message=string}
// @Failure      404  {object}  object{error=string}
// @Router       /api/keys/{id} [delete]
func DeleteAPIKeyHandler(c *gin.Context) {
	userID := uint(c.GetFloat64("user_id"))
	id := c.Param("id")

	result := db.DB.Where("user_id = ? AND id = ?", userID, id).Delete(&models.APIKey{})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "API Key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API Key deleted"})
}

// ToggleAPIKey toggles the active status of an API key.
// @Summary      Toggle API Key
// @Description  Enables or disables a specific API key.
// @Tags         API Keys
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "API Key ID"
// @Success      200  {object}  object{message=string,active=bool}
// @Failure      404  {object}  object{error=string}
// @Router       /api/keys/toggle/{id} [patch]
func ToggleAPIKey(c *gin.Context) {
	userID := uint(c.GetFloat64("user_id"))
	id := c.Param("id")

	var exisintKey models.APIKey

	if err := db.DB.Where("user_id = ? AND id = ?", userID, id).First(&exisintKey).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	exisintKey.Active = !exisintKey.Active
	db.DB.Save(&exisintKey)

	c.JSON(http.StatusOK, gin.H{"message": "API key activation updated", "active": exisintKey.Active})
}
