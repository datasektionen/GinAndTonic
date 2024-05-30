package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	feature_services "github.com/DowLucas/gin-ticket-release/pkg/services/features"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketReleaseController struct {
	DB    *gorm.DB
	trpds *services.TicketReleasePaymentDeadline
}

func NewTicketReleaseController(db *gorm.DB) *TicketReleaseController {
	return &TicketReleaseController{DB: db, trpds: services.NewTicketReleasePaymentDeadline(db)}
}

type TicketReleaseRequest struct {
	EventID                int    `json:"event_id"`
	Name                   string `json:"name"`
	Description            string `json:"description"`
	Open                   int    `json:"open"`
	Close                  int    `json:"close"`
	AllowExternal          bool   `json:"allow_external"`
	TicketReleaseMethodID  int    `json:"ticket_release_method_id"`
	OpenWindowDuration     int    `json:"open_window_duration"`
	MaxTicketsPerUser      int    `json:"max_tickets_per_user"`
	NotificationMethod     string `json:"notification_method"`
	CancellationPolicy     string `json:"cancellation_policy"`
	IsReserved             bool   `json:"is_reserved"`
	PromoCode              string `json:"promo_code"`
	TicketsAvailable       int    `json:"tickets_available"`
	MethodDescription      string `json:"method_description"`
	SaveTemplate           bool   `json:"save_template"`
	PaymentDeadline        string `json:"payment_deadline"`
	ReservePaymentDuration string `json:"reserve_payment_duration"`
	AllocationCutOff       string `json:"allocation_cut_off"`
}

func (trmc *TicketReleaseController) CreateTicketRelease(c *gin.Context) {
	var req TicketReleaseRequest

	user := c.MustGet("user").(models.User)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if there is a ticket release with the same name and event ID
	var checkTicketRelease models.TicketRelease
	if err := trmc.DB.Where("name = ? AND event_id = ?", req.Name, req.EventID).First(&checkTicketRelease).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket release with the same name already exists"})
		return
	}

	var event models.Event
	if err := trmc.DB.First(&event, "id = ?", req.EventID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
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
		MethodDescription:     req.MethodDescription,
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
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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
			hashedPromoCode, err := utils.EncryptString(req.PromoCode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash promo code"})
				return
			}

			promoCode = &hashedPromoCode
		}
	}

	var allocationCutOff *time.Time = nil
	if req.AllocationCutOff != "" {
		t, err := time.Parse("2006-01-02", req.AllocationCutOff)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid allocation cut off: %s", req.AllocationCutOff)})
			return
		}
		allocationCutOff = &t
	}

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
		TicketsAvailable:            req.TicketsAvailable,
		AllowExternal:               req.AllowExternal,
		SaveTemplate:                req.SaveTemplate,
		AllocationCutOff:            allocationCutOff,
	}

	if err := tx.Create(&ticketRelease).Error; err != nil {
		tx.Rollback()
		utils.HandleDBError(c, err, "creating the ticket release")
		return
	}

	if err := trmc.handlePaymentDeadline(req, &ticketRelease, tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ReservePaymentDuration != "" {
		duration, err := time.ParseDuration(req.ReservePaymentDuration)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration"})
			return
		}

		ticketRelease.PaymentDeadline.ReservePaymentDuration = &duration

		if !ticketRelease.PaymentDeadline.Validate(&ticketRelease, &ticketRelease.Event) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment deadline"})
			return
		}

		if err := tx.Save(&ticketRelease.PaymentDeadline).Error; err != nil {
			tx.Rollback()
			utils.HandleDBError(c, err, "updating the ticket release payment deadline")
			return
		}
	}

	eventID := string(req.EventID)                                                                                              // Convert req.EventID to string
	err = feature_services.IncrementFeatureUsage(tx, user.Network.PlanEnrollment.ID, "max_ticket_releases_per_event", &eventID) // Pass the address of eventID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"ticket_release": ticketRelease})
}

func (trmc *TicketReleaseController) ListEventTicketReleases(c *gin.Context) {
	var ticketReleases []models.TicketRelease

	eventID := c.Param("eventID")
	user := c.MustGet("user").(models.User)

	// Convert the event ID to an integer
	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Find the event with the given ID
	// Preload
	if err := trmc.DB.
		Preload("TicketReleaseMethodDetail.ticketReleaseMethod").
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
			if ticketRelease.UserHasAccessToTicketRelease(trmc.DB, user.UGKthID) {
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

func handlePromoCode(req *TicketReleaseRequest, c *gin.Context) (*string, bool) {
	if req.PromoCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Promo code is required for reserved ticket releases"})
		return nil, false
	}

	hashedPromoCode, err := utils.EncryptString(req.PromoCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash promo code"})
		return nil, false
	}

	return &hashedPromoCode, true
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
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)
	var ticketRelease models.TicketRelease

	if err := tx.Preload("PaymentDeadline").Preload("Event").First(&ticketRelease, "event_id = ? AND id = ?", eventIDInt, ticketReleaseIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket release not found"})
		return
	}

	var promoCode *string
	var ok bool

	if (ticketRelease.IsReserved && req.IsReserved) || (!ticketRelease.IsReserved && req.IsReserved) {
		// This means that the ticket release is either reserved and the request is to update the promo code
		// or the ticket release is not reserved and the request is to reserve it
		promoCode, ok = handlePromoCode(&req, c)
		if !ok {
			return
		}
	} else if ticketRelease.IsReserved && !req.IsReserved {
		promoCode = nil
	}

	var allocationCutOff *time.Time = nil
	if req.AllocationCutOff != "" {
		t, err := time.Parse("2006-01-02", req.AllocationCutOff)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid allocation cut off: %s", req.AllocationCutOff)})
			return
		}
		allocationCutOff = &t
	}

	// update
	ticketRelease.Open = int64(req.Open)
	ticketRelease.Close = int64(req.Close)
	ticketRelease.Name = req.Name
	ticketRelease.Description = req.Description
	ticketRelease.TicketsAvailable = req.TicketsAvailable
	ticketRelease.IsReserved = req.IsReserved
	ticketRelease.PromoCode = promoCode
	ticketRelease.AllowExternal = req.AllowExternal
	ticketRelease.SaveTemplate = req.SaveTemplate
	ticketRelease.AllocationCutOff = allocationCutOff

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
	ticketReleaseMethodDetails.TicketReleaseMethodID = uint(req.TicketReleaseMethodID)
	ticketReleaseMethodDetails.MethodDescription = req.MethodDescription

	if err := ticketReleaseMethodDetails.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = trmc.handlePaymentDeadline(req, &ticketRelease, tx)
	if err != nil {
		log.Println(err) // This won't terminate the program
	}

	if err := trmc.handleReservePaymentDuration(req, &ticketRelease, tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := trmc.handleReservePaymentDuration(req, &ticketRelease, tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// ManuallyTryToAllocateReserveTickets manually tries to allocate reserve tickets
func (trmc *TicketReleaseController) ManuallyTryToAllocateReserveTickets(c *gin.Context) {
	var ticketRelease models.TicketRelease

	ticketReleaseID := c.Param("ticketReleaseID")
	eventID := c.Param("eventID")
	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID"})
		return
	}

	eventIDInt, err := strconv.Atoi(eventID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := trmc.DB.First(&ticketRelease, "event_id = ? AND id = ?", eventIDInt, ticketReleaseIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket release not found"})
		return
	}

	if ticketRelease.IsReserved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket release is reserved"})
		return
	}

	// Check that ticket release is closed
	if ticketRelease.Close > time.Now().Unix() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket release is not closed"})
		return
	}

	err = jobs.ManuallyProcessAllocateReserveTicketsJob(trmc.DB, uint(ticketReleaseIDInt))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully tried to allocate reserve tickets"})
}

// Update Payment Deadline
func (trmc *TicketReleaseController) UpdatePaymentDeadline(c *gin.Context) {
	var body types.PaymentDeadlineRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketReleaseIDstring := c.Param("ticketReleaseID")
	ticketReleaseID, err := strconv.Atoi(ticketReleaseIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rerr := trmc.trpds.UpdatePaymentDeadline(ticketReleaseID, body)

	if rerr != nil {
		c.JSON(rerr.StatusCode, gin.H{"error": rerr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated payment deadline"})
}

// Get template ticket releases
func (trmc *TicketReleaseController) GetTemplateTicketReleases(c *gin.Context) {
	var ticketReleases []models.TicketRelease

	user := c.MustGet("user").(models.User)

	var organizations []models.Organization = user.Organizations
	// Get organizations latests events

	eventIds := []int{}
	for _, organization := range organizations {
		var events []models.Event
		if err := trmc.DB.Where("organization_id = ?", organization.ID).Find(&events).Error; err != nil {
			utils.HandleDBError(c, err, "listing the events")
			return
		}

		for _, event := range events {
			eventIds = append(eventIds, int(event.ID))
		}
	}

	if err := trmc.DB.
		Preload("TicketReleaseMethodDetail.TicketReleaseMethod").
		Preload("TicketTypes").
		Preload("PaymentDeadline").
		Where("event_id IN (?) AND save_template = ?", eventIds, true).
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
			if ticketRelease.UserHasAccessToTicketRelease(trmc.DB, user.UGKthID) {
				ticketReleasesFiltered = append(ticketReleasesFiltered, ticketRelease)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"ticket_releases": ticketReleasesFiltered})
}

// Unsave template
func (trmc *TicketReleaseController) UnsaveTemplate(c *gin.Context) {
	ticketReleaseID := c.Param("ticketReleaseID")

	// Convert the ticketRelease ID to an integer
	ticketReleaseIDInt, err := strconv.Atoi(ticketReleaseID)

	user := c.MustGet("user").(models.User)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID"})
		return
	}

	var ticketRelease models.TicketRelease
	if err := trmc.DB.First(&ticketRelease, "id = ?", ticketReleaseIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket release not found"})
		return
	}

	if !ticketRelease.UserHasAccessToTicketRelease(trmc.DB, user.UGKthID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "User does not have access to ticket release"})
		return
	}

	ticketRelease.SaveTemplate = false

	if err := trmc.DB.Save(&ticketRelease).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error updating the ticket release"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ticket_release": ticketRelease})
}

func (trmc *TicketReleaseController) handlePaymentDeadline(req TicketReleaseRequest, ticketRelease *models.TicketRelease, tx *gorm.DB) error {
	fmt.Println("PaymentDeadline: ")
	if req.PaymentDeadline == "" {
		return nil
	}

	fmt.Println("PaymentDeadline: ", req.PaymentDeadline)

	paymentDeadline, err := time.Parse("2006-01-02", req.PaymentDeadline)
	if err != nil {
		return fmt.Errorf("Invalid payment deadline")
	}

	var existingPaymentDeadline models.TicketReleasePaymentDeadline
	err = tx.First(&existingPaymentDeadline, "ticket_release_id = ?", ticketRelease.ID).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create the PaymentDeadline if it doesn't exist
		newPaymentDeadline := models.TicketReleasePaymentDeadline{
			TicketReleaseID:        ticketRelease.ID,
			OriginalDeadline:       paymentDeadline,
			ReservePaymentDuration: nil,
		}
		if err := tx.Create(&newPaymentDeadline).Error; err != nil {
			return err
		}
	} else {
		// Update the PaymentDeadline if it exists
		existingPaymentDeadline.OriginalDeadline = paymentDeadline

		if !existingPaymentDeadline.Validate(ticketRelease, &ticketRelease.Event) {
			return fmt.Errorf("Invalid payment deadline")
		}

		if err := tx.Save(&existingPaymentDeadline).Error; err != nil {
			return err
		}
	}

	return nil
}

func (trmc *TicketReleaseController) handleReservePaymentDuration(req TicketReleaseRequest, ticketRelease *models.TicketRelease, tx *gorm.DB) error {
	if req.ReservePaymentDuration == "" {
		return nil
	}

	duration, err := time.ParseDuration(req.ReservePaymentDuration)
	if err != nil {
		return fmt.Errorf("Invalid duration")
	}

	var existingPaymentDeadline models.TicketReleasePaymentDeadline
	err = tx.First(&existingPaymentDeadline, "ticket_release_id = ?", ticketRelease.ID).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create the PaymentDeadline if it doesn't exist
		newPaymentDeadline := models.TicketReleasePaymentDeadline{
			TicketReleaseID:        ticketRelease.ID,
			OriginalDeadline:       ticketRelease.PaymentDeadline.OriginalDeadline,
			ReservePaymentDuration: &duration,
		}
		if err := tx.Create(&newPaymentDeadline).Error; err != nil {
			return err
		}
	} else {
		// Update the PaymentDeadline if it exists
		existingPaymentDeadline.ReservePaymentDuration = &duration

		if !existingPaymentDeadline.Validate(ticketRelease, &ticketRelease.Event) {
			return fmt.Errorf("Invalid payment deadline")
		}

		if err := tx.Save(&existingPaymentDeadline).Error; err != nil {
			return err
		}
	}

	return nil
}
