package controllers

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketReleaseController struct {
	DB *gorm.DB
}

func NewTicketReleaseController(db *gorm.DB) *TicketReleaseController {
	return &TicketReleaseController{DB: db}
}

type TicketReleaseRequest struct {
	EventID               int    `json:"event_id"`
	Open                  int    `json:"open"`
	Close                 int    `json:"close"`
	TicketReleaseMethodID int    `json:"ticket_release_method_id"`
	OpenWindowDuration    int    `json:"open_window_duration"`
	MaxTicketsPerUser     int    `json:"max_tickets_per_user"`
	NotificationMethod    string `json:"notification_method"`
	CancellationPolicy    string `json:"cancellation_policy"`
}

func (trmc *TicketReleaseController) CreateTicketRelease(c *gin.Context) {
	var req TicketReleaseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	println(req.TicketReleaseMethodID)

	// Get ticket release method from id
	var ticketReleaseMethod models.TicketReleaseMethod
	if err := trmc.DB.First(&ticketReleaseMethod, "id = ?", req.TicketReleaseMethodID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release method ID"})
		return
	}

	ticketReleaseMethodDetails := models.TicketReleaseMethodDetail{
		TicketReleaseMethodID: ticketReleaseMethod.ID,
		OpenWindowDuration:    int64(req.OpenWindowDuration),
		NotificationMethod:    req.NotificationMethod,
		CancellationPolicy:    req.CancellationPolicy,
	}

	if err := ticketReleaseMethodDetails.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cancellation policy"})
		return
	}

	// Start transaction
	tx := trmc.DB.Begin()

	if err := tx.Create(&ticketReleaseMethodDetails).Error; err != nil {
		tx.Rollback()
		utils.HandleDBError(c, err, "creating the ticket release method details")
		return
	}

	method, err := models.NewTicketReleaseConfig(ticketReleaseMethod.MethodName, &ticketReleaseMethodDetails)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := method.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketRelease := models.TicketRelease{
		EventID:                     req.EventID,
		Open:                        int64(req.Open),
		Close:                       int64(req.Close),
		HasAllocatedTickets:         false,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetails.ID,
	}

	if err := tx.Create(&ticketRelease).Error; err != nil {
		tx.Rollback()
		utils.HandleDBError(c, err, "creating the ticket release")
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"ticket_release": ticketRelease})
}


func (trmc *TicketReleaseController) ListEventTicketReleases(c *gin.Context) {
	var ticketReleases []models.TicketRelease

	eventID := c.Param("eventID")

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Find the event with the given ID
	// Preload
	if err := trmc.DB.Preload("TicketReleaseMethodDetail.TicketReleaseMethod").Preload("TicketTypes").Where("event_id = ?", eventIDInt).Find(&ticketReleases).Error; err != nil {
		utils.HandleDBError(c, err, "listing the ticket releases")
		return
	}

	c.JSON(http.StatusOK, ticketReleases)
}

func (trmc *TicketReleaseController) GetTicketRelease(c *gin.Context) {
	var ticketRelease models.TicketRelease

	eventID := c.Param("eventID")
	ticketReleaseID := c.Param("ticketReleaseID")

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID"})
		return
	}

	// Find the event with the given ID
	// Preload
	if err := trmc.DB.Preload("TicketTypes").Where("event_id = ? AND id = ?", eventIDInt, ticketReleaseIDInt).First(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	c.JSON(http.StatusOK, ticketRelease)
}

func (trmc *TicketReleaseController) DeleteTicketRelease(c *gin.Context) {
	var ticketRelease models.TicketRelease

	eventID := c.Param("eventID")
	ticketReleaseID := c.Param("ticketReleaseID")

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID"})
		return
	}

	// Find the event with the given ID
	// Preload
	if err := trmc.DB.Preload("TicketTypes").Where("event_id = ? AND id = ?", eventIDInt, ticketReleaseIDInt).First(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := trmc.DB.Delete(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error deleting the ticket release"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ticket_release": ticketRelease})
}

func (trmc *TicketReleaseController) UpdateTicketRelease(c *gin.Context) {
	var ticketRelease models.TicketRelease

	ticketReleaseID := c.Param("ticketReleaseID")
	eventID := c.Param("eventID")

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)

	if err := trmc.DB.First(&ticketRelease, "event_id = ? AND id = ?", eventIDInt, ticketReleaseIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket release not found"})
		return
	}

	if err := c.ShouldBindJSON(&ticketRelease); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := trmc.DB.Save(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ticket_release": ticketRelease})
}
