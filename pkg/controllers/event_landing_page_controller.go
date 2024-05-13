package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EventLandingPageController struct {
	db      *gorm.DB
	service *services.EventLandingPageService
}

func NewEventLandingPageController(db *gorm.DB) *EventLandingPageController {
	return &EventLandingPageController{
		db:      db,
		service: services.NewEventLandingPageService(db),
	}
}

func (elpc *EventLandingPageController) SaveEventLandingPageEditorState(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}
	fmt.Println("Request Body:", string(bodyBytes))

	rerr := elpc.service.SaveEventLandingPageEditorState(bodyBytes, uint(eventID))
	if rerr != nil {
		c.JSON(rerr.StatusCode, gin.H{"error": rerr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event landing page saved successfully"})
}

func (elpc *EventLandingPageController) SaveEventLandingPage(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var body models.EventLandingPage
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body.EventID = uint(eventID)

	rerr := elpc.service.SaveEventLandingPage(&body)
	if rerr != nil {
		c.JSON(rerr.StatusCode, gin.H{"error": rerr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event landing page saved successfully"})
}

func (elpc *EventLandingPageController) GetEventLandingPage(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var landingPage models.EventLandingPage
	result := elpc.db.Where("event_id = ?", eventID).First(&landingPage)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event landing page not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, landingPage)
}

func (elpc *EventLandingPageController) GetEventLandingPageEditorState(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var landingPage models.EventLandingPage
	result := elpc.db.Where("event_id = ?", eventID).First(&landingPage)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event landing page not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, landingPage.EditorState)
}

type EventLandingPageEnabledBody struct {
	Enabled bool `json:"enabled"`
}

func (elpc *EventLandingPageController) ToggleLandingPageEnabled(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var body EventLandingPageEnabledBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rerr := elpc.service.ToggleLandingPageEnabled(uint(eventID), body.Enabled)
	if rerr != nil {
		c.JSON(rerr.StatusCode, gin.H{"error": rerr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event landing page enabled status updated successfully"})
}
