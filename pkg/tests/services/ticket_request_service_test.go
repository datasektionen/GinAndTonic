package services

import (
	"testing"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/database/factory"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/stretchr/testify/assert"
)

var ticketRequestService *services.TicketRequestService

func SetupTicketRequestDB() {
	// Seed the database with necessary data for testing
	// For example, create TicketRelease, TicketType, User, etc.
	// Create a Event
	event := models.Event{
		Name:        "validEventName",
		Description: "validEventDescription",
		Location:    "validEventLocation",
		// Date is a time.Time type
		Date:           time.Now(),
		OrganizationID: 1,
		CreatedBy:      "validUserUGKthID",
	}

	db.Create(&event)

	// Create ticket release method
	ticketReleaseMethod := factory.NewTicketReleaseMethod(
		string(models.FCFS_LOTTERY),
		"validTicketReleaseMethodDescription",
	)

	db.Create(ticketReleaseMethod)

	ticketReleaseMethodDetail := factory.NewTicketReleaseMethodDetail(
		10,
		"Email",
		"Standard",
		3600, // OpenWindowDuration in seconds
		1,    // Example TicketReleaseMethodID
	)

	db.Create(ticketReleaseMethodDetail)

	// Create ticket types

	// Create a TicketRelease
	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Open:                        uint(time.Now().Unix()) - 1000,
		Close:                       uint(time.Now().Unix()) + 1000,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetail.ID,
	}

	ticketType := factory.NewTicketType(event.ID, "validTicketTypeName", "validTicketTypeDescription", 100, 100, false, 1)

	db.Create(&ticketRelease)
	db.Create(ticketType)

	println(ticketType.ID)

	// Initialize the TicketRequestService
	ticketRequestService = services.NewTicketRequestService(db)
	return
}

// Example test for CreateTicketRequest
func TestCreateTicketRequest(t *testing.T) {
	// Create a ticket request

	ticketRequest := factory.NewTicketRequest(
		1,
		1,
		1,
		"validUserUGKthID",
		false,
	)

	// Test creating a valid ticket request
	err := ticketRequestService.CreateTicketRequest(ticketRequest)
	assert.Nil(t, err)

	// Additional test cases...
}

// Example test for GetTicketRequests
func TestGetTicketRequests(t *testing.T) {
	// Create a ticket request

	ticketRequest := factory.NewTicketRequest(
		1,
		1,
		1,
		"validUserUGKthID",
		false,
	)

	db.Create(&ticketRequest)

	// Test getting ticket requests for a valid user
	ticketRequests, err := ticketRequestService.GetTicketRequests("validUserUGKthID")

	assert.Nil(t, err)

	println(len(ticketRequests))

	// Additional test cases...
}

// Additional tests for other methods...
