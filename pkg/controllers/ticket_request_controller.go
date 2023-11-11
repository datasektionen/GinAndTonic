package controllers

import (
	"net/http"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketRequestController struct {
	DB *gorm.DB
}

func NewTicketRequestController(db *gorm.DB) *TicketRequestController {
	return &TicketRequestController{DB: db}
}

func (trc *TicketRequestController) Create(c *gin.Context) {
	var ticketRequest models.TicketRequest

	UGKthID, _ := c.Get("ugkthid")
	ticketRequest.UserUGKthID = UGKthID.(string)

	// Bind
	if err := c.ShouldBindJSON(&ticketRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !trc.isTicketReleaseOpen(ticketRequest.TicketReleaseID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket release is not open"})
		return
	}

	if !trc.isTicketTypeValid(ticketRequest.TicketTypeID, ticketRequest.TicketReleaseID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket type is not valid for ticket release"})
		return
	}

	if userAlreadyHasATicketToEvent(trc, UGKthID.(string), ticketRequest.TicketReleaseID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already has a ticket to this event"})
		return
	}

	// Create
	if err := trc.DB.Create(&ticketRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error creating the ticket request"})
		return
	}

	c.JSON(http.StatusCreated, ticketRequest)
}

/*
List all ticket request for the user
*/
func (trc *TicketRequestController) Get(c *gin.Context) {
	var ticketRequests []models.TicketRequest

	UGKthID, _ := c.Get("ugkthid")

	if err := trc.DB.Preload("TicketType").Preload("TicketRelease.TicketReleaseMethodDetail").Where("user_ug_kth_id = ?", UGKthID.(string)).Find(&ticketRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error listing the ticket requests"})
		return
	}

	c.JSON(http.StatusOK, ticketRequests)
}

func (trc *TicketRequestController) isTicketReleaseOpen(ticketReleaseID uint) bool {
	var ticketRelease models.TicketRelease
	now := uint(time.Now().Unix())

	if err := trc.DB.Where("id = ?", ticketReleaseID).First(&ticketRelease).Error; err != nil {
		return false
	}
	return now >= ticketRelease.Open && now <= ticketRelease.Close
}

func (trc *TicketRequestController) isTicketTypeValid(ticketTypeID uint, ticketReleaseID uint) bool {
	var count int64
	trc.DB.Model(&models.TicketRelease{}).Joins("JOIN ticket_types ON ticket_types.ticket_release_id = ticket_releases.id").
		Where("ticket_types.id = ? AND ticket_releases.id = ?", ticketTypeID, ticketReleaseID).Count(&count)

	return count > 0
}

func userAlreadyHasATicketToEvent(trc *TicketRequestController, userUGKthID string, ticketReleaseID uint) bool {
	var ticketRelease models.TicketRelease

	// Get event ID from ticket release
	if err := trc.DB.Where("id = ?", ticketReleaseID).First(&ticketRelease).Error; err != nil {
		return false
	}

	var count int64
	trc.DB.Model(&models.TicketRequest{}).
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Joins("JOIN events ON ticket_releases.event_id = events.id").
		Where("ticket_requests.user_ug_kth_id = ? AND events.id = ?", userUGKthID, ticketRelease.EventID).
		Count(&count)

	return count > 0
}
