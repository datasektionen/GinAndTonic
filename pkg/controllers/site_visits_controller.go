package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SitVisitsController struct {
	DB *gorm.DB
}

// NewSitVisitsController creates a new controller with the given database client
func NewSitVisitsController(db *gorm.DB) *SitVisitsController {
	return &SitVisitsController{DB: db}
}

func (svc *SitVisitsController) Get(c *gin.Context) {
	eventID := c.Param("eventID")

	var summaries []models.EventSiteVisitSummary
	// Order from most recent to least recent
	if err := svc.DB.Where("event_id = ?", eventID).Order("created_at desc").Find(&summaries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// total_site_visits, difference from last week

	if len(summaries) == 0 {
		c.JSON(http.StatusOK, gin.H{"total_site_visits": 0})
		return
	}

	totalSiteVisists := summaries[0].TotalVisits

	// Find the last week's summary
	var lastWeekSummary models.EventSiteVisitSummary
	for _, summary := range summaries {
		// Check the created at date
		if summary.CreatedAt.AddDate(0, 0, 7).After(summaries[0].CreatedAt) {
			lastWeekSummary = summary
			break
		}
	}
	// If there is no last week summary just use the last element
	lastWeekSummary = summaries[len(summaries)-1]

	c.JSON(http.StatusOK, gin.H{
		"total_site_visits":             totalSiteVisists,
		"total_site_visits_last_week":   lastWeekSummary.TotalVisits,
		"unique_visitors":               summaries[0].UniqueUsers,
		"unique_visitors_last_week":     lastWeekSummary.UniqueUsers,
		"last_week_date":                lastWeekSummary.CreatedAt,
		"num_ticket_requests":           summaries[0].NumTickets,
		"num_ticket_requests_last_week": lastWeekSummary.NumTickets,
		"total_income":                  summaries[0].TotalIncome,
		"total_income_last_week":        lastWeekSummary.TotalIncome,
	})
}
