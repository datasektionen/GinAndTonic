package middleware

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserLoader(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		UGKthID, exists := c.Get("user_id")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		user, err := models.GetUserByUGKthIDIfExist(db, UGKthID.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Store the user in the context
		c.Set("user", user)

		c.Next()
	}
}
