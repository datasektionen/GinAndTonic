package middleware

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UpdateSiteVisits(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if ?dont_count_site_visit query parameter is set
		if c.Query("dont_count_site_visit") != "" {
			c.Next()
			return
		}

		userID := c.MustGet("ugkthid").(string)
		eventIDstring := c.Param("eventID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
			return
		}

		eventID, err := strconv.Atoi(eventIDstring)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			return
		}

		userAgent := c.GetHeader("User-Agent")
		referrerURL := c.GetHeader("Referer")
		location := c.ClientIP()

		siteVisit := models.EventSiteVisit{
			UserUGKthID: userID,
			UserAgent:   userAgent,
			ReferrerURL: referrerURL,
			Location:    location,
			EventID:     uint(eventID),
		}

		result := db.Create(&siteVisit)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update site visits"})
			return
		}

		c.Next()
	}
}
