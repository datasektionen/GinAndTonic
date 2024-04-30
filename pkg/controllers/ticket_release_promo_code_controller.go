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

func (ctrl *TicketReleasePromoCodeController) GuestCreate(c *gin.Context) {
	ugkthid := c.Param("ugkthid")
	if ugkthid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	requestToken := c.DefaultQuery("request_token", "")
	if requestToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing request token"})
		return
	}

	var user models.User
	if err := ctrl.DB.Where("ug_kth_id = ? AND request_token = ?", ugkthid, requestToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid request token"})
		return
	}

	eventID := c.Param("eventID")
	intEventID, err := strconv.Atoi(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	promoCode := c.DefaultQuery("promo_code", "")

	if promoCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing promo code"})
		return
	}

	// Find ticket release based on promo code
	var ticketReleases []models.TicketRelease
	if err := ctrl.DB.Where("event_id = ?", intEventID).Find(&ticketReleases).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	for _, ticketRelease := range ticketReleases {
		if ticketRelease.PromoCode != nil {
			decryptedPromoCode, err := utils.DecryptString(*ticketRelease.PromoCode)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decrypting promo code"})
				return
			}

			if decryptedPromoCode == promoCode {
				ticketRelease.UserUnlockReservedTicketRelease(&user)

				// Save ticket release
				if err := ctrl.DB.Save(&ticketRelease).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving ticket release"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Guest has successfully unlocked the ticket release!"})
				return
			}
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promo code"})
}

func (ctrl *TicketReleasePromoCodeController) Create(c *gin.Context) {
	eventID := c.Param("eventID")

	intEventID, err := strconv.Atoi(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	ugKthId := c.GetString("user_id")

	var user models.User
	if err := ctrl.DB.Where("ug_kth_id = ?", ugKthId).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user"})
		return
	}

	promoCode := c.DefaultQuery("promo_code", "")

	if promoCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing promo code"})
		return
	}

	// Check if the user is a super admin
	var event models.Event
	if err := ctrl.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Find ticket release based on promo code
	var ticketReleases []models.TicketRelease
	if err := ctrl.DB.Where("event_id = ?", intEventID).Find(&ticketReleases).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	for _, ticketRelease := range ticketReleases {
		if ticketRelease.PromoCode != nil {
			decryptedPromoCode, err := utils.DecryptString(*ticketRelease.PromoCode)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decrypting promo code"})
				return
			}

			if decryptedPromoCode == promoCode {
				ticketRelease.UserUnlockReservedTicketRelease(&user)

				// Save ticket release
				if err := ctrl.DB.Save(&ticketRelease).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving ticket release"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "User has successfully unlocked the ticket release!"})
				return
			}
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promo code"})
}
