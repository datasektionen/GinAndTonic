package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		if roles.(string) != role {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
