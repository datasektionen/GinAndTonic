package controllers

import (
	"net/http"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AllocateTicketsController struct {
	DB                     *gorm.DB
	AllocateTicketsService *services.AllocateTicketsService
}

// NewAllocateTicketsController creates a new controller with the given database client
func NewAllocateTicketsController(db *gorm.DB, ats *services.AllocateTicketsService) *AllocateTicketsController {
	return &AllocateTicketsController{DB: db, AllocateTicketsService: ats}
}

func (atc *AllocateTicketsController) AllocateTickets(c *gin.Context) {
	eventID := c.Param("eventID")
	ticketReleaseID := c.Param("ticketReleaseID")

	var ticketRelease models.TicketRelease

	// Find based on event ID and ticket release ID
	if err := atc.DB.Preload("TicketReleaseMethodDetail.TicketReleaseMethod").Preload("TicketTypes").Where("event_id = ? AND id = ?", eventID, ticketReleaseID).First(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID or ticket release ID"})
		return
	}

	if ticketRelease.HasAllocatedTickets {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tickets already allocated"})
		return
	}

	err := atc.AllocateTicketsService.AllocateTickets(&ticketRelease)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Close the ticket release
	ticketRelease.Close = time.Now().Unix()

	if err := atc.DB.Save(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing ticket release"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tickets allocated"})
}

func (atc *AllocateTicketsController) ListAllocatedTickets(c *gin.Context) {
	ticketReleaseID := c.Param("ticketReleaseID")

	var ticketRequests []models.TicketRequest
	if err := atc.DB.Preload("Tickets").Where("ticket_release_id = ?", ticketReleaseID).Find(&ticketRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error listing the requested tickets"})
		return
	}

	// Get Ticket associated with each RequestedTicket
	var tickets []models.Ticket
	for _, requestedTicket := range ticketRequests {
		tickets = append(tickets, requestedTicket.Tickets...)
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

func (atc *AllocateTicketsController) ListAllocatedTicketsForEvent(c *gin.Context) {
	eventID := c.Param("eventID")

	var tickets []models.Ticket
	if err := atc.DB.
		Preload("TicketRequest").
		Preload("TicketRequest.TicketType").
		Preload("TicketRequest.TicketRelease").
		Joins("JOIN ticket_requests ON tickets.ticket_request_id = ticket_requests.id").
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ?", eventID).
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error listing the requested tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}
