package controllers

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketReleaseController struct {
	DB *gorm.DB
}

func NewTicketReleaseController(db *gorm.DB) *TicketReleaseController {
	return &TicketReleaseController{DB: db}
}

type TicketReleaseRequest struct {
	EventID               int    `json:"event_id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Open                  int    `json:"open"`
	Close                 int    `json:"close"`
	TicketReleaseMethodID int    `json:"ticket_release_method_id"`
	OpenWindowDuration    int    `json:"open_window_duration"`
	MaxTicketsPerUser     int    `json:"max_tickets_per_user"`
	NotificationMethod    string `json:"notification_method"`
	CancellationPolicy    string `json:"cancellation_policy"`
	IsReserved            bool   `json:"is_reserved"`
	PromoCode             string `json:"promo_code"`
}

func (trmc *TicketReleaseController) CreateTicketRelease(c *gin.Context) {
	var req TicketReleaseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get ticket release method from id
	var ticketReleaseMethod models.TicketReleaseMethod
	if err := trmc.DB.First(&ticketReleaseMethod, "id = ?", req.TicketReleaseMethodID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release method ID"})
		return
	}

	ticketReleaseMethodDetails := models.TicketReleaseMethodDetail{
		TicketReleaseMethodID: ticketReleaseMethod.ID,
		OpenWindowDuration:    int64(req.OpenWindowDuration),
		NotificationMethod:    req.NotificationMethod,
		CancellationPolicy:    req.CancellationPolicy,
		MaxTicketsPerUser:     uint(req.MaxTicketsPerUser),
	}

	if err := ticketReleaseMethodDetails.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := trmc.DB.Begin()

	if err := tx.Create(&ticketReleaseMethodDetails).Error; err != nil {
		tx.Rollback()
		utils.HandleDBError(c, err, "creating the ticket release method details")
		return
	}

	method, err := models.NewTicketReleaseConfig(ticketReleaseMethod.MethodName, &ticketReleaseMethodDetails)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := method.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var promoCode *string
	if req.IsReserved {
		if req.PromoCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Promo code is required for reserved ticket releases"})
			return
		} else {
			hashedPromoCode, err := utils.HashString(req.PromoCode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash promo code"})
				return
			}

			promoCode = &hashedPromoCode
		}
	}

	println(req.IsReserved)
	println(*promoCode)

	ticketRelease := models.TicketRelease{
		EventID:                     req.EventID,
		Name:                        req.Name,
		Description:                 req.Description,
		Open:                        int64(req.Open),
		Close:                       int64(req.Close),
		HasAllocatedTickets:         false,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetails.ID,
		IsReserved:                  req.IsReserved,
		PromoCode:                   promoCode,
	}

	if err := tx.Create(&ticketRelease).Error; err != nil {
		tx.Rollback()
		utils.HandleDBError(c, err, "creating the ticket release")
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"ticket_release": ticketRelease})
}

func (trmc *TicketReleaseController) ListEventTicketReleases(c *gin.Context) {
	var ticketReleases []models.TicketRelease
	var user models.User

	eventID := c.Param("eventID")
	ugkthid, _ := c.Get("ugkthid")

	if err := trmc.DB.Where("ug_kth_id = ?", ugkthid).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user"})
		return
	}

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Find the event with the given ID
	// Preload
	if err := trmc.DB.
		Preload("TicketReleaseMethodDetail.TicketReleaseMethod").
		Preload("TicketTypes").
		Where("event_id = ?", eventIDInt).
		Find(&ticketReleases).Error; err != nil {
		utils.HandleDBError(c, err, "listing the ticket releases")
		return
	}

	// Remove ticket releases that have the property IsReserved set to true
	var ticketReleasesFiltered []models.TicketRelease = []models.TicketRelease{}

	for _, ticketRelease := range ticketReleases {
		if !ticketRelease.IsReserved {
			ticketReleasesFiltered = append(ticketReleasesFiltered, ticketRelease)
		} else {
			if ticketRelease.UserHasAccessToTicketRelease(&user) {
				ticketReleasesFiltered = append(ticketReleasesFiltered, ticketRelease)
			}
		}

	}

	c.JSON(http.StatusOK, gin.H{"ticket_releases": ticketReleasesFiltered})
}

func (trmc *TicketReleaseController) GetTicketRelease(c *gin.Context) {
	var ticketRelease models.TicketRelease

	eventID := c.Param("eventID")
	ticketReleaseID := c.Param("ticketReleaseID")

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID"})
		return
	}

	// Find the event with the given ID
	// Preload
	if err := trmc.DB.Preload("TicketTypes").Where("event_id = ? AND id = ?", eventIDInt, ticketReleaseIDInt).First(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if ticketRelease.IsReserved {
		// Get promo_code query string
		promoCode := c.DefaultQuery("promo_code", "")
		println(promoCode)
		if promoCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing promo code"})
			return
		}

		// Hash the promo code
		checked, err := utils.CompareHashAndString(*ticketRelease.PromoCode, promoCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promo code"})
			return
		}

		if !checked {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promo code"})
			return
		}
	}

	c.JSON(http.StatusOK, ticketRelease)
}

func (trmc *TicketReleaseController) DeleteTicketRelease(c *gin.Context) {
	var ticketRelease models.TicketRelease

	eventID := c.Param("eventID")
	ticketReleaseID := c.Param("ticketReleaseID")

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID"})
		return
	}

	// Find the event with the given ID
	// Preload
	if err := trmc.DB.Preload("TicketTypes").Where("event_id = ? AND id = ?", eventIDInt, ticketReleaseIDInt).First(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := trmc.DB.Delete(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error deleting the ticket release"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ticket_release": ticketRelease})
}

func (trmc *TicketReleaseController) UpdateTicketRelease(c *gin.Context) {
	var req TicketReleaseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketReleaseID := c.Param("ticketReleaseID")
	eventID := c.Param("eventID")

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// start transaction
	tx := trmc.DB.Begin()

	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)
	var ticketRelease models.TicketRelease

	if err := tx.First(&ticketRelease, "event_id = ? AND id = ?", eventIDInt, ticketReleaseIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket release not found"})
		return
	}

	// update
	ticketRelease.Open = int64(req.Open)
	ticketRelease.Close = int64(req.Close)

	// Update ticket release method details
	var ticketReleaseMethodDetails models.TicketReleaseMethodDetail
	if err := tx.First(&ticketReleaseMethodDetails, "id = ?", ticketRelease.TicketReleaseMethodDetailID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release method details ID"})
		return
	}

	ticketReleaseMethodDetails.OpenWindowDuration = int64(req.OpenWindowDuration)
	ticketReleaseMethodDetails.NotificationMethod = req.NotificationMethod
	ticketReleaseMethodDetails.CancellationPolicy = req.CancellationPolicy
	ticketReleaseMethodDetails.MaxTicketsPerUser = uint(req.MaxTicketsPerUser)

	if err := ticketReleaseMethodDetails.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cancellation policy"})
		return
	}

	if err := tx.Save(&ticketReleaseMethodDetails).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error updating the ticket release method details"})
		return
	}

	if err := tx.Save(&ticketRelease).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error updating the ticket release"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"ticket_release": ticketRelease})
}
