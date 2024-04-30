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

type EventFormFieldResponseController struct {
	db      *gorm.DB
	service *services.EventFormFieldResponseService
}

func NewEventFormFieldResponseController(db *gorm.DB) *EventFormFieldResponseController {
	service := services.NewEventFormFieldResponseService(db)
	return &EventFormFieldResponseController{db: db, service: service}
}

func (effrc *EventFormFieldResponseController) Upsert(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	_, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	ticketRequestID := c.Param("ticketRequestID")
	user := c.MustGet("user").(models.User)
	var request []types.EventFormFieldResponseCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := effrc.service.Upsert(&user, ticketRequestID, request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (effrc *EventFormFieldResponseController) GuestUpsert(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	_, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	ugkthid := c.Param("ugkthid")
	if ugkthid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing ugkthid"})
		return
	}

	request_token := c.Query("request_token")
	if request_token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing request_token"})
		return
	}

	var user models.User
	if err := effrc.db.
		Where("id = ? AND request_token = ?", ugkthid, request_token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	ticketRequestID := c.Param("ticketRequestID")
	var request []types.EventFormFieldResponseCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := effrc.service.Upsert(&user, ticketRequestID, request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
