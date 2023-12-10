package controllers

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketReleasePromoCodeController struct {
	DB *gorm.DB
}

func NewTicketReleasePromoCodeController(db *gorm.DB) *TicketReleasePromoCodeController {
	return &TicketReleasePromoCodeController{DB: db}
}

func (ctrl *TicketReleasePromoCodeController) Get(c *gin.Context) {
	eventID := c.Param("eventId")
	promoCode := c.DefaultQuery("promo_code", "")

	if promoCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing promo code"})
		return
	}

	userID := c.GetString("ugkthid")

	var userUnlockedTicketRelease models.UserUnlockedTicketRelease
	if err := ctrl.DB.Where("user_id = ? AND ticket_release_id = ?", userID, eventID).First(&userUnlockedTicketRelease).Error; err != nil {
		// An unexpected error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
		return
	}

	// The user has unlocked the ticket release
	c.JSON(http.StatusOK, gin.H{"message": "User has unlocked the ticket release"})
}

func (ctrl *TicketReleasePromoCodeController) Create(c *gin.Context) {
	eventID := c.Param("eventID")

	intEventID, err := strconv.Atoi(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	ugKthId := c.GetString("ugkthid")
	promoCode := c.DefaultQuery("promo_code", "")

	if promoCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing promo code"})
		return
	}

	hashedPromoCode, err := utils.HashString(promoCode)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing promo code"})
		return
	}

	// Check if the user is a super admin
	var event models.Event
	if err := ctrl.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Find ticket release based on promo code
	var ticketRelease models.TicketRelease
	if err := ctrl.DB.Where("event_id = ? AND promo_code = ?", intEventID, hashedPromoCode).First(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promo code"})
		return
	}

	ticketReleasePromoCode := models.UserUnlockedTicketRelease{
		TicketReleaseID: ticketRelease.ID,
		UserUGKthID:     ugKthId,
	}

	// Check if the user has already unlocked the ticket release
	var userUnlockedTicketRelease models.UserUnlockedTicketRelease
	if err := ctrl.DB.Where("user_ug_kth_id = ? AND ticket_release_id = ?", ugKthId, ticketRelease.ID).First(&userUnlockedTicketRelease).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "User has already unlocked the ticket release"})
		return
	}

	if err := ctrl.DB.Create(&ticketReleasePromoCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has successfully unlocked the ticket release!"})
}
