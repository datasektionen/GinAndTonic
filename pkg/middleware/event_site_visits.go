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

		refID := c.Param("refID")
		eventIDstring := c.Param("eventID")
		var eventID uint

		if refID == "" && eventIDstring == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing reference ID or event ID"})
			return
		}

		if refID != "" && eventIDstring == "" {
			var event models.Event
			if result := db.Where("reference_id = ?", refID).First(&event); result.Error != nil {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Event not found"})
				return
			}

			eventID = event.ID
		} else if refID == "" && eventIDstring != "" {
			eventIDint, err := strconv.Atoi(eventIDstring)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
				return
			}

			eventID = uint(eventIDint)
		}

		userAgent := c.GetHeader("User-Agent")
		referrerURL := c.GetHeader("Referer")
		location := c.ClientIP()

		siteVisit := models.EventSiteVisit{
			UserAgent:   userAgent,
			ReferrerURL: referrerURL,
			Location:    location,
			EventID:     eventID,
		}

		result := db.Create(&siteVisit)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update site visits"})
			return
		}

		c.Next()
	}
}
