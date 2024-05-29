package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handles big request where we will create event, ticket release and ticket types.

type CompleteEventWorkflowController struct {
	DB      *gorm.DB
	service *services.CompleteEventWorkflowService
}

// NewCompleteEventWorkflowcontroller creates a new controller with the given database client
func NewCompleteEventWorkflowController(db *gorm.DB) *CompleteEventWorkflowController {
	return &CompleteEventWorkflowController{
		DB:      db,
		service: services.NewCompleteEventWorkflowService(db),
	}
}

func (cewc *CompleteEventWorkflowController) CreateEvent(c *gin.Context) {
	var err error

	user := c.MustGet("user").(models.User)

	var req types.EventFullWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("EventFullWorkflowRequest", req)
	event, err := cewc.service.CreateEvent(req, &user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func (cewc *CompleteEventWorkflowController) CreateTicketRelease(c *gin.Context) {
	var err error

	var req types.TicketReleaseFullWorkFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ugkthid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = cewc.service.CreateTicketRelease(req, eventID, ugkthid.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Ticket release created"})
}
