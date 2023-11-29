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

func (suite *TicketsTestSuite) SetupSuite() {
	db, err := testutils.SetupTestDatabase(false)
	suite.Require().NoError(err)

	suite.ts = services.NewTicketService(db)
	suite.ats = services.NewAllocateTicketsService(db)
	suite.db = db
}

func (suite *TicketsTestSuite) SetupTest() {
	testutils.SetupEventWorkflow(suite.db)
}

func (suite *TicketsTestSuite) TearDownTest() {
	testutils.CleanupTestDatabase(suite.db)
}

func (suite *TicketsTestSuite) TestGetAllTickets() {
	// Get ticket release
	var ticketRelease models.TicketRelease
	if err := suite.db.First(&ticketRelease).Error; err != nil {
		suite.FailNow(err.Error())
	}

	// Create ticket requests
	requests := 1000

	for i := 0; i < requests; i++ {
		testutils.CreateTicketRequestWorkflow(suite.db, ticketRelease.ID, time.Now())
	}

	// Allocate
	suite.ats.AllocateTickets(&ticketRelease)

	// Get tickets
	tickets, err := suite.ts.GetAllTickets(int(ticketRelease.EventID))
	suite.NoError(err)

	println("len(tickets): ", len(tickets))
}

func TestTicketsTestSuite(t *testing.T) {
	suite.Run(t, new(TicketsTestSuite))
}
