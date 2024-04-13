package test_service

import (
	"os"
	"testing"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/tests/testutils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AllocateReserveTicketsTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *services.AllocateTicketsService
}

func (suite *AllocateReserveTicketsTestSuite) SetupTest() {
	os.Setenv("ENV", "test")
	db, err := testutils.SetupTestDatabase(false)
	suite.Require().NoError(err)

	suite.db = db
	suite.service = services.NewAllocateTicketsService(db)
}

func (suite *AllocateReserveTicketsTestSuite) TearDownTest() {
	testutils.CleanupTestDatabase(suite.db)
}

func (suite *AllocateReserveTicketsTestSuite) createTicketRelease(totalTickets int, methodName models.TRM, openWindowDuration int64, openTime int64) models.TicketRelease {
	return models.TicketRelease{
		TicketTypes: []models.TicketType{
			{
				EventID:         1,
				Name:            "Standard",
				Description:     "Standard ticket",
				Price:           420,
				TicketReleaseID: 1,
			},
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

func (suite *AllocateReserveTicketsTestSuite) createAndSaveTicketRequests(tr models.TicketRelease, requests int, timeOffset time.Duration) {
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: tr.ID,
			TicketTypeID:    1,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(timeOffset * time.Duration(i)),
			},
		}
		suite.db.Create(&req)
	}
}

func (suite *AllocateReserveTicketsTestSuite) validateTicketAllocation(expectedAllocated, expectedReserve, expectedUnhandled int) {
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

func (suite *AllocateReserveTicketsTestSuite) TestNotAllocatedReserveTicketsWhenNotRemovedTickets() {
	var err error

	service := services.NewAllocateTicketsService(suite.db)

	totalTickets := 100
	requests := 120

	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketTypeID:    1,
			TicketReleaseID: tr.ID,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		err := suite.db.Create(&req)
		suite.NoError(err.Error)
	}

	var ticketRequests []models.TicketRequest
	suite.db.Where("ticket_release_id = ?", tr.ID).Find(&ticketRequests)
	suite.Equal(requests, len(ticketRequests))

	err = service.AllocateTickets(&tr, nil)
	suite.NoError(err)

	//Validate
	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	// Now we run the job that allocates the reserve tickets
	err = jobs.AllocateReserveTicketsJob(suite.db)
	suite.NoError(err)

	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateReserveTicketsTestSuite) TestAllocateReserveTicketsWhenRemovedTickets() {
	var err error

	service := services.NewAllocateTicketsService(suite.db)

	totalTickets := 100
	requests := 120
	num_tickets_to_remove := 20

	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: tr.ID,
			TicketTypeID:    1,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		err := suite.db.Create(&req)
		suite.NoError(err.Error)
	}

	var ticketRequests []models.TicketRequest
	suite.db.Where("ticket_release_id = ?", tr.ID).Find(&ticketRequests)
	suite.Equal(requests, len(ticketRequests))

	err = service.AllocateTickets(&tr, nil)
	suite.NoError(err)

	//Validate
	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	// We now cancel 2 tickets
	var normalTickets []models.Ticket
	suite.db.Where("is_reserve = ?", false).Find(&normalTickets)

	for i := 0; i < num_tickets_to_remove; i++ {
		suite.db.Delete(&normalTickets[i])
	}

	// Now we run the job that allocates the reserve tickets
	err = jobs.AllocateReserveTicketsJob(suite.db)
	suite.NoError(err)

	suite.validateTicketAllocation(totalTickets, 0, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateReserveTicketsTestSuite) TestAllocatedTicketsOnTicketsThatHasntPaidInTime() {
	var err error

	service := services.NewAllocateTicketsService(suite.db)

	totalTickets := 100
	requests := 120
	num_tickets_to_modify := 20

	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: tr.ID,
			TicketTypeID:    1,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		err := suite.db.Create(&req)
		suite.NoError(err.Error)
	}

	var ticketRequests []models.TicketRequest
	suite.db.Where("ticket_release_id = ?", tr.ID).Find(&ticketRequests)
	suite.Equal(requests, len(ticketRequests))

	err = service.AllocateTickets(&tr, nil)
	suite.NoError(err)

	//Validate
	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	// We now cancel 2 tickets
	var normalTickets []models.Ticket
	suite.db.Where("is_reserve = ?", false).Find(&normalTickets)

	// Set the 20 first tickets to not paid
	for i := 0; i < num_tickets_to_modify; i++ {
		// Set updated at to 25 hours in the past
		suite.db.Model(&normalTickets[i]).Update("updated_at", time.Now().Add(time.Duration(-25)*time.Hour))
	}

	// Now we run the job that allocates the reserve tickets
	err = jobs.AllocateReserveTicketsJob(suite.db)
	suite.NoError(err)

	suite.validateTicketAllocation(totalTickets, 0, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateReserveTicketsTestSuite) TestNotAllocatedTicketsOnTicketsThatHasPaidInTime() {
	var err error

	service := services.NewAllocateTicketsService(suite.db)

	totalTickets := 100
	requests := 120
	num_tickets_to_modify := 20

	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: tr.ID,
			TicketTypeID:    1,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		err := suite.db.Create(&req)
		suite.NoError(err.Error)
	}

	var ticketRequests []models.TicketRequest
	suite.db.Where("ticket_release_id = ?", tr.ID).Find(&ticketRequests)
	suite.Equal(requests, len(ticketRequests))

	err = service.AllocateTickets(&tr, nil)
	suite.NoError(err)

	//Validate
	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	// We now cancel 2 tickets
	var normalTickets []models.Ticket
	suite.db.Where("is_reserve = ?", false).Find(&normalTickets)

	// Set the 20 first tickets to be paid within time
	for i := 0; i < num_tickets_to_modify; i++ {
		// Set updated at to 23 hours in the past
		suite.db.Model(&normalTickets[i]).Update("updated_at", time.Now().Add(time.Duration(-23)*time.Hour))
	}

	// Now we run the job that allocates the reserve tickets
	err = jobs.AllocateReserveTicketsJob(suite.db)
	suite.NoError(err)

	// Validate that the ticket allocation remains the same
	suite.validateTicketAllocation(totalTickets, requests-totalTickets, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateReserveTicketsTestSuite) TestAllocateTicketsWithNoRequests() {
	var err error

	service := services.NewAllocateTicketsService(suite.db)

	totalTickets := 100
	requests := 0

	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: tr.ID,
			TicketTypeID:    1,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		err := suite.db.Create(&req)
		suite.NoError(err.Error)
	}

	var ticketRequests []models.TicketRequest
	suite.db.Where("ticket_release_id = ?", tr.ID).Find(&ticketRequests)
	suite.Equal(requests, len(ticketRequests))

	err = service.AllocateTickets(&tr, nil)
	suite.Error(err)

	err = jobs.AllocateReserveTicketsJob(suite.db)
	suite.NoError(err)

	//Validate
	suite.validateTicketAllocation(0, 0, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateReserveTicketsTestSuite) TestAllocateTicketsWithEqualTicketsAndRequests() {
	var err error

	service := services.NewAllocateTicketsService(suite.db)

	totalTickets := 100
	requests := 100

	tr := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)

	if err := suite.db.Create(&tr).Error; err != nil {
		panic(err)
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: tr.ID,
			TicketTypeID:    1,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		err := suite.db.Create(&req)
		suite.NoError(err.Error)
	}

	var ticketRequests []models.TicketRequest
	suite.db.Where("ticket_release_id = ?", tr.ID).Find(&ticketRequests)
	suite.Equal(requests, len(ticketRequests))

	err = service.AllocateTickets(&tr, nil)
	suite.NoError(err)

	err = jobs.AllocateReserveTicketsJob(suite.db)
	suite.NoError(err)

	//Validate
	suite.validateTicketAllocation(totalTickets, 0, 0)

	suite.Equal(tr.HasAllocatedTickets, true)
}

func (suite *AllocateReserveTicketsTestSuite) TestAllocateTicketsWithMultipleTicketReleases() {
	var err error

	service := services.NewAllocateTicketsService(suite.db)

	totalTickets := 100
	requests := 120
	num_tickets_to_remove := 20

	tr1 := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)
	tr2 := suite.createTicketRelease(totalTickets, models.FCFS_LOTTERY, time.Now().Unix()-1000, time.Now().Unix()-2000)

	if err := suite.db.Create(&tr1).Error; err != nil {
		panic(err)
	}

	if err := suite.db.Create(&tr2).Error; err != nil {
		panic(err)
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req1 := models.TicketRequest{
			TicketReleaseID: tr1.ID,
			TicketTypeID:    1,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		req2 := models.TicketRequest{
			TicketReleaseID: tr2.ID,
			TicketTypeID:    1,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		suite.db.Create(&req2)
		suite.db.Create(&req1)

	}

	var ticketRequests1, ticketRequests2 []models.TicketRequest
	suite.db.Where("ticket_release_id = ?", tr1.ID).Find(&ticketRequests1)
	suite.db.Where("ticket_release_id = ?", tr2.ID).Find(&ticketRequests2)
	suite.Equal(requests, len(ticketRequests1))
	suite.Equal(requests, len(ticketRequests2))

	err = service.AllocateTickets(&tr1, nil)
	suite.NoError(err)

	err = service.AllocateTickets(&tr2, nil)
	suite.NoError(err)

	// Remove tickets from each ticket release
	var normalTickets1, normalTickets2 []models.Ticket
	// Use join to get the ticket.ticket_request.ticket_release_id to get the correct tickets
	// And where is_reserve = false to get the normal tickets
	suite.db.Joins("JOIN ticket_requests ON ticket_requests.id = tickets.ticket_request_id").Where("ticket_requests.ticket_release_id = ? AND tickets.is_reserve = ?", tr1.ID, false).Find(&normalTickets1)
	suite.db.Joins("JOIN ticket_requests ON ticket_requests.id = tickets.ticket_request_id").Where("ticket_requests.ticket_release_id = ? AND tickets.is_reserve = ?", tr2.ID, false).Find(&normalTickets2)

	for i := 0; i < num_tickets_to_remove; i++ {
		suite.db.Delete(&normalTickets1[i])
		suite.db.Delete(&normalTickets2[i])
	}

	err = jobs.AllocateReserveTicketsJob(suite.db)
	suite.NoError(err)

	var reserveTickets1, reserveTickets2 []models.Ticket
	suite.db.Where("is_reserve = ?", true).Find(&reserveTickets1)
	suite.db.Where("is_reserve = ?", true).Find(&reserveTickets2)

	suite.Equal(len(reserveTickets1), 0)
	suite.Equal(len(reserveTickets2), 0)

	suite.Equal(tr1.HasAllocatedTickets, true)
	suite.Equal(tr2.HasAllocatedTickets, true)
}
func TestAllocateReserveTicketsTestSuite(t *testing.T) {
	suite.Run(t, new(AllocateReserveTicketsTestSuite))
}
