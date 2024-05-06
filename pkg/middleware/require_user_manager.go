package middleware

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
)

/*
	RequireUserManager is a middleware that checks if the user is a manager.
*/

func RequireUserManager() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.MustGet("user").(models.User)

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		if user.IsSuperAdmin() {
			c.Next()
			return
		}

		if !user.IsEventManager() {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is not a manager"})
			c.Abort()
			return
		}

		c.Next()
	}
}
