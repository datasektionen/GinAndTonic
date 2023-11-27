package test_service

import (
	"fmt"
	"testing"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/database/factory"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/tests/testutils"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TicketRequestTestSuite struct {
	suite.Suite
	db                   *gorm.DB
	ticketRequestService *services.TicketRequestService
}

func (suite *TicketRequestTestSuite) SetupTest() {
	var err error
	db, err := testutils.SetupTestDatabase()
	suite.Require().NoError(err)

	suite.db = db
	suite.ticketRequestService = services.NewTicketRequestService(db)
}

func (suite *TicketRequestTestSuite) TearDownTest() {
	testutils.CleanupTestDatabase(suite.db)
}

func SetupTicketRequestDB(db *gorm.DB) {
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
}

func (suite *TicketRequestTestSuite) TestCreateTicketRequest() {
	SetupTicketRequestDB(suite.db) // Assuming this function is adapted to accept a *gorm.DB parameter

	ticketRequest := factory.NewTicketRequest(
		1,
		1,
		1,
		"validUserUGKthID",
		false,
	)

	err := suite.ticketRequestService.CreateTicketRequest(ticketRequest)
	suite.Nil(err)

	var ticketRequestFromDB []models.TicketRequest

	ticketRequestFromDB, err = suite.ticketRequestService.GetTicketRequests("validUserUGKthID")
	suite.Nil(err)
	suite.Equal(1, len(ticketRequestFromDB))
}

func (suite *TicketRequestTestSuite) TestCreateTicketRequestAboveMaxTicketsPerUser() {
	SetupTicketRequestDB(suite.db) // Assuming this function is adapted to accept a *gorm.DB parameter

	// ticketRequestService := services.NewTicketRequestService(suite.db)

	for i := 0; i < 10; i++ {
		ticketRequest := factory.NewTicketRequest(
			1,
			1,
			1,
			"validUserUGKthID",
			false,
		)

		err := suite.ticketRequestService.CreateTicketRequest(ticketRequest)
		suite.Nil(err)
	}

	// Create one more ticket request
	ticketRequest := factory.NewTicketRequest(
		1,
		1,
		1,
		"validUserUGKthID",
		false,
	)

	err := suite.ticketRequestService.CreateTicketRequest(ticketRequest)
	suite.NotNil(err)

	var ticketRequestFromDB []models.TicketRequest
	ticketRequestFromDB, err = suite.ticketRequestService.GetTicketRequests("validUserUGKthID")

	suite.Nil(err)
	suite.Equal(10, len(ticketRequestFromDB))

	suite.Equal(10, len(ticketRequestFromDB))
}

func (suite *TicketRequestTestSuite) TestCreateTicketRequestAfterOpenWindowDurationIsReserveTicket() {
	// Create event with ticket release method detail with open window duration of 1 second

	event := models.Event{
		Name:        "validEventName",
		Description: "validEventDescription",
		Location:    "validEventLocation",
		// Date is a time.Time type
		Date:           time.Now(),
		OrganizationID: 1,
		CreatedBy:      "validUserUGKthID",
	}

	suite.db.Create(&event)

	// Create ticket release method
	ticketReleaseMethod := factory.NewTicketReleaseMethod(
		string(models.FCFS_LOTTERY),
		"validTicketReleaseMethodDescription",
	)

	suite.db.Create(ticketReleaseMethod)

	ticketReleaseMethodDetail := factory.NewTicketReleaseMethodDetail(
		10,
		"Email",
		"Standard",
		10, // OpenWindowDuration in seconds
		1,  // Example TicketReleaseMethodID
	)

	suite.db.Create(ticketReleaseMethodDetail)

	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Open:                        uint(time.Now().Unix()) - 11, // Ticket is created 1 second after open window duration has passed which means it is a reserve ticket
		Close:                       uint(time.Now().Unix()) + 20,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetail.ID,
	}

	suite.db.Create(&ticketRelease)

	ticketType := factory.NewTicketType(event.ID, "validTicketTypeName", "validTicketTypeDescription", 100, 100, false, 1)

	suite.db.Create(ticketType)

	// Create ticket request
	ticketRequest := factory.NewTicketRequest(
		1,
		1,
		1,
		"validUserUGKthID",
		false,
	)

	err := suite.ticketRequestService.CreateTicketRequest(ticketRequest)
	suite.Nil(err)

	// Check that it is a reserve ticket
	var ticketRequestFromDB []models.TicketRequest
	ticketRequestFromDB, err = suite.ticketRequestService.GetTicketRequests("validUserUGKthID")

	suite.Nil(err)
	suite.Equal(1, len(ticketRequestFromDB))

	requested_time := ticketRequestFromDB[0].CreatedAt
	window_close_time := utils.ConvertUNIXTimeToDateTime(int64(ticketRelease.Open + ticketReleaseMethodDetail.OpenWindowDuration))

	fmt.Printf("Requested time: %s\n", requested_time.Format(time.RFC3339))
	fmt.Printf("Window close time: %s\n", window_close_time.Format(time.RFC3339))

	suite.Equal(true, requested_time.After(window_close_time))
}

func TestTicketRequestTestSuite(t *testing.T) {
	suite.Run(t, new(TicketRequestTestSuite))
}
