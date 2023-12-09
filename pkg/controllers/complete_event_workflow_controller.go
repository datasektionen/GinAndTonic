package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handles big request where we will create event, ticket release and ticket types.

type CompleteEventWorkflowController struct {
	DB      *gorm.DB
	service *services.EventService
}

// NewCompleteEventWorkflowcontroller creates a new controller with the given database client
func NewCompleteEventWorkflowController(db *gorm.DB, service *services.EventService) *CompleteEventWorkflowController {
	return &CompleteEventWorkflowController{
		DB:      db,
		service: services.NewEventService(db),
	}
}

func (cewc *CompleteEventWorkflowController) CreateEvent(c *gin.Context) {
	var err error

	var req types.EventFullWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ugkthid, exists := c.Get("ugkthid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = cewc.service.CreateEvent(req, ugkthid.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Event created"})
}
