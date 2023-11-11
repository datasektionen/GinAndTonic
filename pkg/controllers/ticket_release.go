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
	TicketRelease              models.TicketRelease             `json:"ticket_release"`
	TicketReleaseMethodDetails models.TicketReleaseMethodDetail `json:"ticket_release_method_details"`
}

func (trmc *TicketReleaseController) CreateTicketRelease(c *gin.Context) {
	var req TicketReleaseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketRelease := req.TicketRelease
	ticketReleaseMethodDetails := req.TicketReleaseMethodDetails

	// Get ticket release method from id
	var ticketReleaseMethod models.TicketReleaseMethod
	if err := trmc.DB.First(&ticketReleaseMethod, ticketReleaseMethodDetails.TicketReleaseMethodID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release method ID"})
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

	// Start transaction
	tx := trmc.DB.Begin()

	// Create ticket release method details
	if err := tx.Create(&ticketReleaseMethodDetails).Error; err != nil {
		tx.Rollback()
		utils.HandleDBError(c, err, "creating the ticket release method details")
		return
	}

	ticketRelease.TicketReleaseMethodDetailID = ticketReleaseMethodDetails.ID

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
