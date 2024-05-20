package middleware

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeEventAccess(db *gorm.DB, requiredRole models.OrgRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		ugkthid, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		eventID := c.Param("eventID")
		eventIDInt, err := utils.ParseStringToUint(eventID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			c.Abort()
			return
		}

		event, err := models.GetEvent(db, uint(eventIDInt))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch event"})
			c.Abort()
			return
		}

		authorized, err := CheckUserAuthorization(db, uint(event.OrganizationID), ugkthid.(string), requiredRole)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check authorization"})
			c.Abort()
			return
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Event not found"})
			c.Abort()
			return
		}

		c.Set("event", event) // Store the event in the context for later use
		c.Next()
	}
}
