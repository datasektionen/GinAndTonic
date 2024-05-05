package feature_middleware

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
)

func RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user's package tier
		user := c.MustGet("user").(models.User)

		// Check if the user is an super_admin
		if user.IsSuperAdmin() {
			c.Next()
			return
		}

		if !user.IsEventManager() {
			c.JSON(403, gin.H{"error": "User is not an event manager"})
			c.Abort()
			return
		}

	}
}
