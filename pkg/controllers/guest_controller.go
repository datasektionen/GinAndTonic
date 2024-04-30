package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GuestController struct {
	DB *gorm.DB
}

// NewGuestController creates a new controller with the given database client
func NewGuestController(db *gorm.DB) *GuestController {
	return &GuestController{DB: db}
}

func (gc *GuestController) Get(c *gin.Context) {
	request_token := c.Query("request_token")

	if request_token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing request_token"})
		return
	}

	ugkthid := c.Param("ugkthid")
	if ugkthid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing ugkthid"})
		return
	}

	var user models.User
	if err := gc.DB.
		Preload("Roles").
		Preload("TicketRequests.TicketRelease.Event.FormFields").
		Preload("TicketRequests.TicketType").
		Preload("TicketRequests.TicketAddOns.AddOn").
		Preload("TicketRequests.EventFormReponses").
		Preload("TicketRequests.Tickets").
		Where("id = ? AND request_token = ?", ugkthid, request_token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
