package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	feature_services "github.com/DowLucas/gin-ticket-release/pkg/services/features"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/pkg/validation"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EventController struct {
	DB      *gorm.DB
	service *services.EventService
}

// NewEventController creates a new controller with the given database client
func NewEventController(db *gorm.DB) *EventController {
	return &EventController{DB: db, service: services.NewEventService(db)}
}

// CreateEvent handles the creation of an event
func (ec *EventController) CreateEvent(c *gin.Context) {
	// Use types.EventRequest instead of models.Event
	var eventRequest types.EventRequest

	if err := c.ShouldBindJSON(&eventRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ugkthid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check that organization exists
	var organization models.Organization
	if err := ec.DB.First(&organization, eventRequest.OrganizationID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	event := models.Event{
		Name:           eventRequest.Name,
		Description:    eventRequest.Description,
		Location:       eventRequest.Location,
		Date:           eventRequest.Date,
		OrganizationID: eventRequest.OrganizationID,
		IsPrivate:      eventRequest.IsPrivate,
		CreatedBy:      ugkthid.(string),
	}

	if event.IsPrivate {
		token, err := utils.GenerateSecretToken()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error generating the secret token"})
			return
		}

		event.SecretToken = token
	}

	if err := ec.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error creating the event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func (ec *EventController) ListEvents(c *gin.Context) {
	var events []models.Event
	user := c.MustGet("user").(models.User)

	// Pagination
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset value"})
		return
	}

	// Filtering
	name := c.Query("name")

	// Sorting
	sort := c.DefaultQuery("sort", "created_at desc")

	query := ec.DB.Limit(limit).Offset(offset).Order(sort)

	// Apply filtering if a name is provided
	if name != "" {
		query = query.Where("name = ?", name)
	}

	query.Find(&events)

	ugkthid := c.GetString("user_id")

	var authorizedEvents []models.Event
	for _, event := range events {
		if event.IsPrivate {
			if user.IsSuperAdmin() {
				authorizedEvents = append(authorizedEvents, event)
			} else {
				authorized, err := middleware.CheckUserAuthorization(ec.DB,
					uint(event.OrganizationID),
					ugkthid,
					models.OrganizationMember)

				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check authorization"})
					return
				}

				if authorized {
					authorizedEvents = append(authorizedEvents, event)
				}
			}
		} else {
			authorizedEvents = append(authorizedEvents, event)
		}
	}

	c.JSON(http.StatusOK, authorizedEvents)
}

func (ec *EventController) GetEvent(c *gin.Context) {
	_, ugkthid_exists := c.Get("user_id")

	if ugkthid_exists {
		ec.UserGetEvent(c)
	} else {
		ec.CustomerGetEvent(c)
	}
}

func (ec *EventController) UserGetEvent(c *gin.Context) {
	var event models.Event
	var user models.User

	eventID := c.Param("eventID")
	ugkthid, _ := c.Get("user_id")

	if ugkthid != nil {
		if err := ec.DB.Where("id = ?", ugkthid.(string)).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
	}

	err := ec.DB.
		Preload("Organization").
		Preload("TicketReleases").
		Preload("TicketReleases.TicketTypes").
		Preload("TicketReleases.ReservedUsers").
		Preload("TicketReleases.Event").
		Preload("TicketReleases.TicketReleaseMethodDetail.TicketReleaseMethod").
		Preload("TicketReleases.PaymentDeadline").
		Preload("FormFields").
		Preload("LandingPage").
		Preload("TicketReleases.AddOns").
		Find(&event, eventID).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if the event is private
	authorized := false

	if event.IsPrivate {
		if user.UGKthID != "" && user.IsSuperAdmin() {
			authorized = true
		} else {
			var err error
			authorized, err = middleware.CheckUserAuthorization(ec.DB,
				uint(event.OrganizationID),
				user.UGKthID,
				models.OrganizationMember)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check authorization"})
				return
			}

			// Get the secret token from the request
			secretToken := c.Query("secret_token")

			// Check the secret token
			if secretToken == event.SecretToken {
				authorized = true
			}
		}
	} else {
		authorized = true
	}

	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Event not found"})
		return
	}

	for i, ticketRelease := range event.TicketReleases {
		if ticketRelease.UserHasAccessToTicketRelease(ec.DB, user.UGKthID) {
			if ticketRelease.PromoCode != nil {
				decryptedPromoCode, err := utils.DecryptString(*ticketRelease.PromoCode)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error decrypting the promo code"})
					return
				}
				event.TicketReleases[i].PromoCode = &decryptedPromoCode
			}
		}
	}

	// Remove ticket releases that have the property IsReserved set to true
	var ticketReleasesFiltered []models.TicketRelease = []models.TicketRelease{}

	for _, ticketRelease := range event.TicketReleases {
		if !ticketRelease.IsReserved {
			ticketReleasesFiltered = append(ticketReleasesFiltered, ticketRelease)
		} else {
			if ticketRelease.UserHasAccessToTicketRelease(ec.DB, user.UGKthID) {
				ticketReleasesFiltered = append(ticketReleasesFiltered, ticketRelease)
			}
		}
	}

	event.TicketReleases = ticketReleasesFiltered

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    gin.H{"event": event, "timestamp": time.Now().Unix()},
		"message": "Event retrieved successfully",
	})
}

// GetEvent handles retrieving an event by ID
func (ec *EventController) CustomerGetEvent(c *gin.Context) {
	var event models.Event

	refID := c.Param("refID")
	promoCodes := c.QueryArray("promo_codes")

	err := ec.DB.
		Preload("Organization").
		Preload("TicketReleases").
		Preload("TicketReleases.TicketTypes").
		Preload("TicketReleases.ReservedUsers").
		Preload("TicketReleases.Event").
		Preload("TicketReleases.TicketReleaseMethodDetail.TicketReleaseMethod").
		Preload("TicketReleases.PaymentDeadline").
		Preload("FormFields").
		Preload("TicketReleases.AddOns").
		Preload("LandingPage").
		Where("reference_id = ?", refID).
		First(&event).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if the event is private
	authorized := false

	if event.IsPrivate {
		secretToken := c.Query("secret_token")
		if secretToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Event not found"})
			return
		}

		// Check the secret token
		if secretToken == event.SecretToken {
			authorized = true
		}
	} else {
		authorized = true
	}

	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Event not found"})
		return
	}

	// Remove duplicates from the promoCodes array
	promoCodes = utils.RemoveDuplicates(promoCodes)

	// Remove ticket releases that have the property IsReserved set to true
	var ticketReleasesFiltered []models.TicketRelease = []models.TicketRelease{}

	for _, ticketRelease := range event.TicketReleases {
		if !ticketRelease.IsReserved {
			ticketReleasesFiltered = append(ticketReleasesFiltered, ticketRelease)
		} else {
			// Check if any of the promo codes matches
			if ticketRelease.HasPromoCode() {
				for _, promoCode := range promoCodes {
					decryptedPromoCode, err := utils.DecryptString(*ticketRelease.PromoCode)

					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error decrypting the promo code"})
						return
					}

					if promoCode == decryptedPromoCode {
						ticketReleasesFiltered = append(ticketReleasesFiltered, ticketRelease)
					}
				}
			}
		}
	}

	event.TicketReleases = ticketReleasesFiltered

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    gin.H{"event": event},
		"message": "Event retrieved successfully",
	})
}

func (ec *EventController) GetTimestamp(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    gin.H{"timestamp": time.Now().Unix()},
		"message": "Timestamp retrieved successfully",
	})
}

// UpdateEvent handles updating an event by ID
func (ec *EventController) UpdateEvent(c *gin.Context) {
	var eventRequest types.EventRequest
	var event models.Event
	id := c.Param("eventID")

	if err := ec.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if err := c.ShouldBindJSON(&eventRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.Name = eventRequest.Name
	event.Description = eventRequest.Description
	event.Location = eventRequest.Location
	event.Date = eventRequest.Date
	if eventRequest.EndDate != nil {
		endDate := time.Unix(*eventRequest.EndDate, 0)
		event.EndDate = &endDate
	}
	event.OrganizationID = eventRequest.OrganizationID
	event.IsPrivate = eventRequest.IsPrivate

	if eventRequest.IsPrivate && event.SecretToken == "" {
		token, err := utils.GenerateSecretToken()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error generating the secret token"})
			return
		}

		event.SecretToken = token
	}

	if err := ec.DB.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error updating the event"})
		return
	}

	if err := validation.ValidateEventDates(ec.DB, event.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event": event})
}

// DeleteEvent handles deleting an event by ID
func (ec *EventController) DeleteEvent(c *gin.Context) {
	var event models.Event
	id := c.Param("eventID")
	user := c.MustGet("user").(models.User)

	if err := ec.DB.Delete(&event, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Decrement the max_events
	err := feature_services.DecrementFeatureUsage(ec.DB, user.Network.PlanEnrollment.ID, "max_events", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error decrementing the feature usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"timestamp": time.Now().Unix(),
	})
}

func (ec *EventController) ListTickets(c *gin.Context) {
	eventID := c.Param("eventID")

	var tickets []models.Ticket
	if err := ec.DB.
		Unscoped().
		Preload("User.FoodPreferences").
		Preload("TicketType").
		Preload("EventFormReponses.EventFormField").
		Preload("TicketOrder.TicketRelease.TicketReleaseMethodDetail.TicketReleaseMethod").
		Preload("TicketAddOns.AddOn").
		Preload("TicketOrder.Order.Details").
		Joins("JOIN ticket_orders ON tickets.ticket_order_id = ticket_orders.id").
		Joins("JOIN ticket_releases ON ticket_orders.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ?", eventID).
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "There was an error retrieving the tickets",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    gin.H{"tickets": tickets},
		"message": "Tickets retrieved successfully",
	})
}

// GetEventSecretToken returns the secret token for an event
func (ec *EventController) GetEventSecretToken(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := ec.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.SecretToken == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event does not have a secret token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secret_token": event.SecretToken})
}

func (ec *EventController) GetUsersView(c *gin.Context) {
	// A user should be able to view the event landing page, so get the html and css

	eventIDstring := c.Param("refID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	result := ec.DB.Preload("LandingPage").First(&event, eventID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event landing page not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if event.LandingPage.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event landing page not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"html": event.LandingPage.HTML,
		"css":  event.LandingPage.CSS,
	})
}
