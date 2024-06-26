package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketTypeController struct {
	DB *gorm.DB
}

func NewTicketTypeController(db *gorm.DB) *TicketTypeController {
	return &TicketTypeController{DB: db}
}

func (ttc *TicketTypeController) ListAllTicketTypes(c *gin.Context) {
	var ticketTypes []models.TicketType

	if err := ttc.DB.Find(&ticketTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error listing the ticket types"})
		return
	}

	c.JSON(http.StatusOK, ticketTypes)
}

// Check that event exists before creating ticket release

func (ttc *TicketTypeController) CreateTicketTypes(c *gin.Context) {
	var ticketTypes []models.TicketType

	if err := c.ShouldBindJSON(&ticketTypes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := ttc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for idx, ticketType := range ticketTypes {
		// Check that event exists
		if !checkEventExists(ttc, ticketType.EventID) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID for ticket type at index " + strconv.Itoa(idx)})
			return
		}

		// Check that ticket release exists
		if !checkTicketReleaseExists(ttc, ticketType.TicketReleaseID) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID for ticket type at index " + strconv.Itoa(idx)})
			return
		}

		if err := tx.Create(&ticketType).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusCreated, ticketTypes)
}

// Get ticket types from event id and ticket release id
func (ttc *TicketTypeController) GetEventTicketTypes(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := parseIntParam(eventIDstring, "eventID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketReleaseIDstring := c.Param("ticketReleaseID")
	ticketReleaseID, err := parseIntParam(ticketReleaseIDstring, "ticketReleaseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ticketTypes []models.TicketType

	if err := ttc.DB.Where("event_id = ? AND ticket_release_id = ?", eventID, ticketReleaseID).Find(&ticketTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error getting the ticket types"})
		return
	}

	c.JSON(http.StatusOK, ticketTypes)
}

// Update ticket types from event id and ticket release id
func (ttc *TicketTypeController) UpdateEventTicketTypes(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := parseIntParam(eventIDstring, "eventID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketReleaseIDstring := c.Param("ticketReleaseID")
	ticketReleaseID, err := parseIntParam(ticketReleaseIDstring, "ticketReleaseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var allTicketTypes []models.TicketType
	if err := ttc.DB.Where("event_id = ? AND ticket_release_id = ?", eventID, ticketReleaseID).Find(&allTicketTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error getting the ticket types"})
		return
	}

	var ticketTypes []models.TicketType

	if err := c.ShouldBindJSON(&ticketTypes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := ttc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get all ticket types for the event and ticket release
	var existingTicketTypes []models.TicketType
	if err := tx.Where("event_id = ? AND ticket_release_id = ?", eventID, ticketReleaseID).Find(&existingTicketTypes).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, ticketType := range ticketTypes {
		// Check that event exists
		if !checkEventExists(ttc, ticketType.EventID) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID for ticket type"})
			return
		}

		// Check that ticket release exists
		if !checkTicketReleaseExists(ttc, ticketType.TicketReleaseID) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID for ticket type"})
			return
		}

		if !checkTicketTypeExists(ttc, ticketType.ID) {
			// Create the ticket type
			if err := tx.Create(&ticketType).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		updateData := map[string]interface{}{
			"price":       ticketType.Price,
			"name":        ticketType.Name,
			"description": ticketType.Description,
		}

		// Try to update the ticket type
		if err := tx.Model(&ticketType).
			Where("event_id = ? AND ticket_release_id = ? AND id = ?", eventID, ticketReleaseID, ticketType.ID).
			Updates(updateData).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	// Remove ticket types that are not in the request
	for _, existingTicketType := range existingTicketTypes {
		found := false
		for _, ticketType := range ticketTypes {
			if existingTicketType.ID == ticketType.ID {
				found = true
				break
			}
		}
		if !found {
			if err := tx.Delete(&existingTicketType).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, ticketTypes)
}

func checkEventExists(ttc *TicketTypeController, eventID uint) bool {
	var event models.Event

	if err := ttc.DB.First(&event, eventID).Error; err != nil {
		return false
	}

	return true
}

func checkTicketReleaseExists(ttc *TicketTypeController, ticketReleaseID uint) bool {
	var ticketRelease models.TicketRelease

	if err := ttc.DB.First(&ticketRelease, ticketReleaseID).Error; err != nil {
		return false
	}

	return true
}

func checkTicketTypeExists(ttc *TicketTypeController, ticketTypeID uint) bool {
	var ticketType models.TicketType

	if err := ttc.DB.First(&ticketType, ticketTypeID).Error; err != nil {
		return false
	}

	return true
}

func parseIntParam(param string, paramName string) (int, error) {
	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s: %s", paramName, param)
	}

	return value, nil
}
