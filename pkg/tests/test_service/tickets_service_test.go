package test_service

import (
	"testing"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/tests/testutils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TicketsTestSuite struct {
	suite.Suite
	db  *gorm.DB
	ts  *services.TicketService
	ats *services.AllocateTicketsService
}

func (suite *TicketsTestSuite) SetupTest() {
	db, err := testutils.SetupTestDatabase(false)
	suite.Require().NoError(err)

	suite.ts = services.NewTicketService(db)
	suite.ats = services.NewAllocateTicketsService(db)

	suite.db = db
	testutils.SetupEventWorkflow(suite.db)
}

func (suite *TicketsTestSuite) TearDownTest() {
	testutils.CleanupTestDatabase(suite.db)
}

func (suite *TicketsTestSuite) TestGetAllTickets() {
	// Get ticket release
	var ticketRelease models.TicketRelease
	if err := suite.db.Preload("TicketReleaseMethodDetail.TicketReleaseMethod").First(&ticketRelease).Error; err != nil {
		suite.Fail("Failed to get ticket release")
	}

	// Create ticket requests
	requests := 1000

	for i := 0; i < requests; i++ {
		testutils.CreateTicketRequestWorkflow(suite.db, ticketRelease.ID, time.Now())
	}

	// Allocate tickets
	err := suite.ats.AllocateTickets(&ticketRelease, nil)
	suite.Require().NoError(err)

	// Get all tickets
	tickets, err := suite.ts.GetAllTicketsToEvent(int(ticketRelease.EventID))
	suite.Require().NoError(err)

	suite.Equal(requests, len(tickets))
}

func (suite *TicketsTestSuite) TestGetTicketOfEvent() {
	// Get ticket release
	var ticketRelease models.TicketRelease
	if err := suite.db.Preload("TicketReleaseMethodDetail.TicketReleaseMethod").First(&ticketRelease).Error; err != nil {
		suite.Fail("Failed to get ticket release")
	}

	// Create ticket requests
	requests := 1000

	for i := 0; i < requests; i++ {
		testutils.CreateTicketRequestWorkflow(suite.db, ticketRelease.ID, time.Now())
	}

	// Allocate tickets
	err := suite.ats.AllocateTickets(&ticketRelease, nil)
	suite.Require().NoError(err)

	// Get ticket
	ticket, err := suite.ts.GetTicketToEvent(int(ticketRelease.EventID), 1)

	suite.Equal(ticket.ID, uint(1))
}

func TestTicketsTestSuite(t *testing.T) {
	suite.Run(t, new(TicketsTestSuite))
}
