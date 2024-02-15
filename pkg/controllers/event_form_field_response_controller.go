package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EventFormFieldResponseController struct {
	service *services.EventFormFieldResponseService
}

func NewEventFormFieldResponseController(db *gorm.DB) *EventFormFieldResponseController {
	service := services.NewEventFormFieldResponseService(db)
	return &EventFormFieldResponseController{service: service}
}

func (effrc *EventFormFieldResponseController) Upsert(c *gin.Context) {
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
