package authentication

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireRole(requiredRole models.RoleType, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Get role from the database
		var requiredRoleDB models.Role
		if err := db.Where("name = ?", requiredRole).First(&requiredRoleDB).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch role"})
			c.Abort()
			return
		}

		// Check if the user has the required role
		hasRequiredRole := false
		for _, role := range roles.([]string) {
			var userRole models.Role
			if err := db.Where("name = ?", role).First(&userRole).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch role"})
				c.Abort()
				return
			}

			if userRole.ID == requiredRoleDB.ID {
				hasRequiredRole = true
				break
			}
		}

		// If the user does not have the required role, they are not authorized
		if !hasRequiredRole {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
