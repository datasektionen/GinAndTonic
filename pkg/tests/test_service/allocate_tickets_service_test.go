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

type AllocateTicketsTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *services.AllocateTicketsService
}

func (suite *AllocateTicketsTestSuite) SetupTest() {
	db, err := testutils.SetupTestDatabase(false)
	suite.Require().NoError(err)

	suite.db = db
	suite.service = services.NewAllocateTicketsService(db)
}

func (suite *AllocateTicketsTestSuite) TearDownTest() {
	testutils.CleanupTestDatabase(suite.db)
}

func (suite *AllocateTicketsTestSuite) createTicketRelease(totalTickets int, methodName models.TRM, openWindowDuration int64, openTime int64) models.TicketRelease {
	return models.TicketRelease{
		TicketTypes: []models.TicketType{
			{},
		},
		TicketReleaseMethodDetail: models.TicketReleaseMethodDetail{
			TicketReleaseMethod: models.TicketReleaseMethod{
				MethodName: string(methodName),
			},
			OpenWindowDuration: openWindowDuration,
		},
		TicketsAvailable: totalTickets,
		Open:             openTime,
	}
}

func (suite *AllocateTicketsTestSuite) createAndSaveTicketRequests(tr models.TicketRelease, requests int, timeOffset time.Duration) {
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: tr.ID,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(timeOffset * time.Duration(i)),
			},
		}
		suite.db.Create(&req)
	}
}

func (suite *AllocateTicketsTestSuite) validateTicketAllocation(expectedAllocated, expectedReserve, expectedUnhandled int) {
	var allocatedTickets, reserveTickets []models.Ticket
	var ticketRequestFromDB []models.TicketRequest

	suite.db.Where("is_reserve = ?", false).Find(&allocatedTickets)
	suite.db.Where("is_reserve = ?", true).Find(&reserveTickets)
	suite.db.Where("is_handled = ?", false).Find(&ticketRequestFromDB)

	suite.Equal(expectedAllocated, len(allocatedTickets))
	suite.Equal(expectedReserve, len(reserveTickets))
	suite.Equal(expectedUnhandled, len(ticketRequestFromDB))

	// Check that the reserve tickets are numbered 1 to len(reserveTickets)
	if len(reserveTickets) > 0 {
		for i, ticket := range reserveTickets {
			suite.Equal(uint(i+1), ticket.ReserveNumber)
		}
	}
}

func (suite *AllocateTicketsTestSuite) TestAllocateTickets() {
	var err error

	service := services.NewAllocateTicketsService(suite.db)

	totalTickets := 1000
	requests := 1200
	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: tr.ID,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		err := suite.db.Create(&req)
		suite.NoError(err.Error)
	}

	// Check that we have 1200 ticket requests
	var ticketRequests []models.TicketRequest
	suite.db.Where("ticket_release_id = ?", tr.ID).Find(&ticketRequests)
	suite.Equal(requests, len(ticketRequests))

	err = service.AllocateTickets(&tr)
	suite.NoError(err)

	// Validate
	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateTicketsTestSuite) TestAllocateTicketsNoRequestsDuringOpenWindow() {
	var err error

	// Create AllocateTicketsService
	ats := services.NewAllocateTicketsService(suite.db)

	var totalTickets int = 100

	// Create a TicketRelease with OpenWindowDuration
	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, 30, time.Now().Unix()-1000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	requests := 1000

	suite.createAndSaveTicketRequests(tr, requests, 100)

	// Allocate tickets
	err = ats.AllocateTickets(&tr)

	// Validate
	suite.NoError(err)

	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateTicketsTestSuite) TestAllocateTicketsNoRequestsAfterOpenWindow() {
	var err error

	// Create AllocateTicketsService
	ats := services.NewAllocateTicketsService(suite.db)

	var totalTickets int = 100

	// Create a TicketRelease with OpenWindowDuration
	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, 30, time.Now().Unix()-1000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	requests := 1000

	suite.createAndSaveTicketRequests(tr, requests, -100)

	// Allocate tickets
	err = ats.AllocateTickets(&tr)

	// Validate
	suite.NoError(err)

	// We should have 100 tickets allocated
	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateTicketsTestSuite) TestCannotAllocateTicketsTwice() {
	var err error

	// Create AllocateTicketsService
	ats := services.NewAllocateTicketsService(suite.db)

	var totalTickets int = 1

	// Create a TicketRelease with OpenWindowDuration
	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, 30, time.Now().Unix()-1000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	requests := 1

	suite.createAndSaveTicketRequests(tr, requests, -100)

	err = ats.AllocateTickets(&tr)
	suite.NoError(err)

	// We should have 100 tickets allocated
	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	suite.Equal(tr.HasAllocatedTickets, true)

	// Try to allocate tickets again
	err = ats.AllocateTickets(&tr)
	suite.Error(err)
}

func TestAllocateTicketsTestSuite(t *testing.T) {
	suite.Run(t, new(AllocateTicketsTestSuite))
}
