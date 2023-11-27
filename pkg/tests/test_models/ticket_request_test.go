package models_test

import (
	"testing"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TicketRequestTestSuite struct {
	suite.Suite
	db *gorm.DB
}

// SetupTest runs before each test in the suite.
func (suite *TicketRequestTestSuite) SetupTest() {
	db, err := testutils.SetupTestDatabase()
	suite.Require().NoError(err)

	suite.db = db
	// Perform setup that needs to happen before each test here
	// e.g., clean database, create necessary records
	testutils.SetupOrganizationWorkflow(suite.db)
	testutils.SetupEventWorkflow(suite.db)
}

// TearDownTest runs after each test in the suite.
func (suite *TicketRequestTestSuite) TearDownTest() {
	testutils.CleanupTestDatabase(suite.db)
}

func (suite *TicketRequestTestSuite) createTicketRequest(ticketAmount int, ticketReleaseID uint, userUGKthID string, ticketTypeID int, isHandled bool) models.TicketRequest {
	// Create a TicketRequest
	ticketRequest := models.TicketRequest{
		TicketAmount:    ticketAmount,
		TicketReleaseID: ticketReleaseID,
		TicketTypeID:    uint(ticketTypeID),
		UserUGKthID:     userUGKthID,
		IsHandled:       isHandled,
	}

	suite.db.Create(&ticketRequest)

	return ticketRequest
}

func (suite *TicketRequestTestSuite) TestCreateTicketRequest() {
	// Create and save a TicketRequest
	ticketRequest := suite.createTicketRequest(1, 1, "validUserUGKthID", 1, false)

	// Assert no error and correct data
	assert.NotZero(suite.T(), ticketRequest.ID)
}

func (suite *TicketRequestTestSuite) TestRetrieveTicketRequest() {
	// Retrieve a TicketRequest by ID
	ticketRequest := suite.createTicketRequest(1, 1, "validUserUGKthID", 1, false)

	// Get the TicketRequest
	var retrievedTicketRequest models.TicketRequest
	err := suite.db.First(&retrievedTicketRequest, ticketRequest.ID).Error
	assert.NoError(suite.T(), err)
}

func (suite *TicketRequestTestSuite) TestUpdateTicketRequest() {
	// Update a TicketRequest
	ticketRequest := suite.createTicketRequest(1, 1, "validUserUGKthID", 1, false)

	var retrievedTicketRequest models.TicketRequest
	err := suite.db.First(&retrievedTicketRequest, ticketRequest.ID).Error
	assert.NoError(suite.T(), err)

	retrievedTicketRequest.TicketAmount = 2

	err = suite.db.Save(&retrievedTicketRequest).Error

	// Assert no error and updated data
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, retrievedTicketRequest.TicketAmount)
}

func (suite *TicketRequestTestSuite) TestDeleteTicketRequest() {
	// Delete a TicketRequest
	suite.createTicketRequest(1, 1, "validUserUGKthID", 1, false)

	var ticketRequest models.TicketRequest
	suite.db.First(&ticketRequest, 1) // Use an existing ID

	err := suite.db.Delete(&ticketRequest).Error

	// Assert no error
	assert.NoError(suite.T(), err)
}

func (suite *TicketRequestTestSuite) TestCascadeOnTicketReleaseDeletedSoft() {
	// Create a TicketRequest
	suite.createTicketRequest(1, 1, "validUserUGKthID", 1, false)

	// Delete ticket release with id 1
	err := models.DeleteTicketRelease(suite.db, 1)

	if err != nil {
		println(err.Error())
	}

	// Assert no error
	assert.NoError(suite.T(), err)

	var retrievedTicketRelease models.TicketRelease
	err = suite.db.Where("id = ?", 1).First(&retrievedTicketRelease).Error

	println(retrievedTicketRelease.ID)

	// Assert error
	assert.Error(suite.T(), err)

	// Try to retrieve the TicketRequest
	var tr models.TicketRequest
	err = suite.db.First(&tr).Error

	// Assert error
	assert.Error(suite.T(), err)
}

func TestTicketRequestTestSuite(t *testing.T) {
	suite.Run(t, new(TicketRequestTestSuite))
}
