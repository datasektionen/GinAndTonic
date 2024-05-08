package controllers

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SendOutController struct {
	DB  *gorm.DB
	sos *services.SendOutService
}

func NewSendOutController(db *gorm.DB) *SendOutController {
	return &SendOutController{DB: db, sos: services.NewSendOutService(db)}
}

func (sor *SendOutController) GetEventSendOuts(c *gin.Context) {
	eventIdString := c.Param("eventID")
	eventId, e := strconv.Atoi(eventIdString)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}

	var sendOuts []models.SendOut
	if err := sor.DB.Preload("Notifications.User").Where("event_id = ?", eventId).Find(&sendOuts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"send_outs": sendOuts})
}

func (sor *SendOutController) SendOut(c *gin.Context) {
	var req types.SendOutRequest

	eventIdString := c.Param("eventID")
	eventId, e := strconv.Atoi(eventIdString)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}

	var event models.Event
	if err := sor.DB.Preload("Organization").Where("id = ?", eventId).First(&event).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get all ticket releases
	var ticketReleases []models.TicketRelease
	if err := sor.DB.Where("id IN ?", req.TicketReleaseIDs).Find(&ticketReleases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err := sor.sos.SendOutEmails(&event, req.Subject, req.Message, ticketReleases, req.Filters)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(200, gin.H{"message": "Emails sent successfully"})
}
