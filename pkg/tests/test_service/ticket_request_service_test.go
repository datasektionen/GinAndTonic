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
	db, err := testutils.SetupTestDatabase(false)
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
	// Creat user
	testutils.SetupOrganizationWorkflow(db)
	// Create a Event

	event := testutils.CreateEventWorkflow(db)

	// Create ticket release method
	testutils.CreateTicketReleaseMethodWorkflow(db)

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
		Open:                        time.Now().Unix() - 1000,
		Close:                       time.Now().Unix() + 1000,
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
		time.Now(),
	)

	_, err := suite.ticketRequestService.CreateTicketRequests([]models.TicketRequest{*ticketRequest})
	suite.Nil(err)

	var ticketRequestFromDB []models.TicketRequest

	ticketRequestFromDB, err = suite.ticketRequestService.GetTicketRequestsForUser("validUserUGKthID", nil)
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
			time.Now(),
		)

		_, err := suite.ticketRequestService.CreateTicketRequests([]models.TicketRequest{*ticketRequest})
		suite.Nil(err)
	}

	// Create one more ticket request
	ticketRequest := factory.NewTicketRequest(
		1,
		1,
		1,
		"validUserUGKthID",
		false,
		time.Now(),
	)

	_, err := suite.ticketRequestService.CreateTicketRequests([]models.TicketRequest{*ticketRequest})
	suite.NotNil(err)

	var ticketRequestFromDB []models.TicketRequest
	ticketRequestFromDB, err = suite.ticketRequestService.GetTicketRequestsForUser("validUserUGKthID", nil)

	suite.Nil(err)
	suite.Equal(10, len(ticketRequestFromDB))

	suite.Equal(10, len(ticketRequestFromDB))
}

func (suite *TicketRequestTestSuite) TestCreateTicketRequestAfterOpenWindowDurationIsReserveTicket() {
	SetupTicketRequestDB(suite.db) // Assuming this function is adapted to accept a *gorm.DB parameter

	var ticketRelease models.TicketRelease
	suite.db.Preload("TicketReleaseMethodDetail.TicketReleaseMethod").First(&ticketRelease)

	ticketReleaseMethodDetail := ticketRelease.TicketReleaseMethodDetail

	// Create ticket request
	ticketRequest := factory.NewTicketRequest(
		1,
		1,
		1,
		"validUserUGKthID",
		false,
		// CreatedAt is one day after the open window duration
		time.Now().Add(time.Duration(ticketReleaseMethodDetail.OpenWindowDuration+86400)*time.Second),
	)

	if err := suite.db.Create(&ticketRequest).Error; err != nil {
		panic(err)
	}

	// Check that it is a reserve ticket
	var ticketRequestFromDB []models.TicketRequest
	ticketRequestFromDB, err := suite.ticketRequestService.GetTicketRequestsForUser("validUserUGKthID", nil)

	suite.Nil(err)
	suite.Equal(1, len(ticketRequestFromDB))

	requested_time := ticketRequestFromDB[0].CreatedAt
	window_close_time := utils.ConvertUNIXTimeToDateTime(int64(ticketRelease.Open + ticketReleaseMethodDetail.OpenWindowDuration))

	fmt.Printf("Requested time: %s\n", requested_time.Format(time.RFC3339))
	fmt.Printf("Window close time: %s\n", window_close_time.Format(time.RFC3339))

	suite.Equal(true, requested_time.After(window_close_time))
}

func (suite *TicketRequestTestSuite) TestCreateTicketRequestAfterClose() {
	// Create event with ticket release method detail with open window duration of 1 second

	event := testutils.CreateEventWorkflow(suite.db)

	suite.db.Create(&event)

	// Create ticket release method
	ticketReleaseMethod := testutils.CreateTicketReleaseMethodWorkflow(suite.db)

	suite.db.Create(ticketReleaseMethod)

	ticketReleaseMethodDetail := testutils.CreateTicketReleaseMethodDetailWorkflow(suite.db)

	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Open:                        time.Now().Unix() - 20,
		Close:                       time.Now().Unix() - 10,
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
		time.Now(),
	)

	_, err := suite.ticketRequestService.CreateTicketRequests([]models.TicketRequest{*ticketRequest})
	suite.NotNil(err)

	// Check that it is a reserve ticket
	var ticketRequestFromDB []models.TicketRequest
	ticketRequestFromDB, err = suite.ticketRequestService.GetTicketRequestsForUser("validUserUGKthID", nil)

	suite.Nil(err)
	suite.Equal(0, len(ticketRequestFromDB))
}

func (suite *TicketRequestTestSuite) TestCreateTicketRequestBeforeOpen() {
	// Create event with ticket release method detail with open window duration of 1 second

	event := testutils.CreateEventWorkflow(suite.db)

	// Create ticket release method
	testutils.CreateTicketReleaseMethodWorkflow(suite.db)

	ticketReleaseMethodDetail := testutils.CreateTicketReleaseMethodDetailWorkflow(suite.db)

	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Open:                        time.Now().Unix() + 10,
		Close:                       time.Now().Unix() + 20,
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
		time.Now(),
	)

	_, err := suite.ticketRequestService.CreateTicketRequests([]models.TicketRequest{*ticketRequest})
	suite.NotNil(err)

	// Check that it is a reserve ticket
	var ticketRequestFromDB []models.TicketRequest
	ticketRequestFromDB, err = suite.ticketRequestService.GetTicketRequestsForUser("validUserUGKthID", nil)

	suite.Nil(err)
	suite.Equal(0, len(ticketRequestFromDB))
}

func (suite *TicketRequestTestSuite) TestUserMakesTwoTicketRequestUnderMaxTicketAmount() {
	event := testutils.CreateEventWorkflow(suite.db)

	// Create ticket release method
	testutils.CreateTicketReleaseMethodWorkflow(suite.db)

	ticketReleaseMethodDetail := testutils.CreateTicketReleaseMethodDetailWorkflow(suite.db)

	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Open:                        time.Now().Unix() - 10,
		Close:                       time.Now().Unix() + 20,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetail.ID,
	}

	suite.db.Create(&ticketRelease)

	ticketType := factory.NewTicketType(event.ID, "validTicketTypeName", "validTicketTypeDescription", 100, 100, false, 1)

	suite.db.Create(ticketType)

	for i := 0; i < 2; i++ {
		// Create ticket request
		ticketRequest := factory.NewTicketRequest(
			5,
			1,
			1,
			"validUserUGKthID",
			false,
			time.Now(),
		)

		err := suite.db.Create(&ticketRequest).Error
		suite.Nil(err)
	}

	ticketRequest := factory.NewTicketRequest(
		1,
		1,
		1,
		"validUserUGKthID",
		false,
		time.Now(),
	)

	_, err := suite.ticketRequestService.CreateTicketRequests([]models.TicketRequest{*ticketRequest})
	suite.NotNil(err)

	// Check that it is a reserve ticket
	var ticketRequestFromDB []models.TicketRequest
	ticketRequestFromDB, err2 := suite.ticketRequestService.GetTicketRequestsForUser("validUserUGKthID", nil)

	suite.Nil(err2)
	suite.Equal(2, len(ticketRequestFromDB))
}

func TestTicketRequestTestSuite(t *testing.T) {
	suite.Run(t, new(TicketRequestTestSuite))
}
