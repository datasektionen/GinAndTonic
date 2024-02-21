package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/pkg/validation"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EventController struct {
	DB *gorm.DB
}

// NewEventController creates a new controller with the given database client
func NewEventController(db *gorm.DB) *EventController {
	return &EventController{DB: db}
}

// CreateEvent handles the creation of an event
func (ec *EventController) CreateEvent(c *gin.Context) {
	// Use types.EventRequest instead of models.Event
	var eventRequest types.EventRequest

	if err := c.ShouldBindJSON(&eventRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ugkthid, exists := c.Get("ugkthid")
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
		Date:           time.Unix(eventRequest.Date, 0),
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

	ugkthid := c.GetString("ugkthid")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get user"})
		return
	}

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

// GetEvent handles retrieving an event by ID
func (ec *EventController) GetEvent(c *gin.Context) {
	var event models.Event
	var user models.User

	id := c.Param("eventID")
	ugkthid, _ := c.Get("ugkthid")

	if err := ec.DB.Where("ug_kth_id = ?", ugkthid.(string)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := ec.DB.
		Preload("Organization").
		Preload("TicketReleases").
		Preload("TicketReleases.TicketTypes").
		Preload("TicketReleases.ReservedUsers").
		Preload("TicketReleases.Event").
		Preload("TicketReleases.TicketReleaseMethodDetail.TicketReleaseMethod").
		First(&event, id).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Check if the event is private
	authorized := false

	if event.IsPrivate {
		if user.IsSuperAdmin() {
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
			fmt.Println("secretToken", secretToken)
			fmt.Println(secretToken == event.SecretToken)

			// Check the secret token
			if secretToken == event.SecretToken {
				authorized = true
			}
		}
	} else {
		authorized = true
	}

	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "User not authorized for this event"})
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

	c.JSON(http.StatusOK, gin.H{"event": event, "timestamp": time.Now().Unix()})
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
	event.Date = time.Unix(eventRequest.Date, 0)
	event.OrganizationID = eventRequest.OrganizationID
	event.IsPrivate = eventRequest.IsPrivate

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

	if err := ec.DB.Delete(&event, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

func (ec *EventController) ListTickets(c *gin.Context) {
	eventID := c.Param("eventID")

	var tickets []models.Ticket
	var ticketRequests []models.TicketRequest
	if err := ec.DB.
		Unscoped().
		Preload("Transaction").
		Preload("User.FoodPreferences").
		Preload("TicketRequest.TicketType").
		Preload("TicketRequest.TicketRelease.TicketReleaseMethodDetail.TicketReleaseMethod").
		Joins("JOIN ticket_requests ON tickets.ticket_request_id = ticket_requests.id").
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ?", eventID).
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error listing the requested tickets"})
		return
	}

	if err := ec.DB.
		Unscoped().
		Preload("User.FoodPreferences").
		Preload("TicketType").
		Preload("TicketRelease.TicketReleaseMethodDetail.TicketReleaseMethod").
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ? AND NOT EXISTS (SELECT 1 FROM tickets WHERE tickets.ticket_request_id = ticket_requests.id)", eventID).
		Find(&ticketRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error listing the requested ticket requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets, "ticket_requests": ticketRequests})
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
