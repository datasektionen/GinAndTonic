package authentication

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireRole(requiredRole string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("role")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Get role from the database
		var userRole models.Role
		var roleString = roles.(string)

		if err := db.Where("name = ?", roleString).First(&userRole).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch role"})
			c.Abort()
			return
		}

		var requiredRoleDB models.Role
		if err := db.Where("name = ?", requiredRole).First(&requiredRoleDB).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch role"})
			c.Abort()
			return
		}

		// Check if the ID of the userRole is less than or equal to the ID of the roleDB
		// If it is not, the user is not authorized
		if userRole.ID > requiredRoleDB.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
