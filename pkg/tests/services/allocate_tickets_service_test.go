package services

import (
	"testing"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAllocateTickets(t *testing.T) {
	service := services.NewAllocateTicketsService(db)

	totalTickets := 1000
	requests := 1200
	ticketRelease := models.TicketRelease{
		TicketTypes: []models.TicketType{
			{
				QuantityTotal: uint(totalTickets),
			},
		},
		TicketReleaseMethodDetail: models.TicketReleaseMethodDetail{
			TicketReleaseMethod: models.TicketReleaseMethod{
				MethodName: string(models.FCFS_LOTTERY),
			},
			OpenWindowDuration: uint(time.Now().Unix()) - 1000,
		},
		Open: uint(time.Now().Unix()) - 2000,
	}

	// Mock ticket requests
	for i := 0; i < requests; i++ {
		req := models.TicketRequest{
			TicketReleaseID: ticketRelease.ID,
			Model: gorm.Model{
				CreatedAt: time.Now().Add(time.Duration(-i) * time.Second),
			}}
		db.Create(&req)
	}

	err := service.AllocateTickets(ticketRelease)
	assert.Nil(t, err)

	// Validate
	var allocatedTickets []models.Ticket
	db.Where("is_reserve = ?", false).Find(&allocatedTickets)
	assert.Equal(t, totalTickets, len(allocatedTickets))

	var reserveTickets []models.Ticket
	db.Where("is_reserve = ?", true).Find(&reserveTickets)
	assert.Equal(t, requests-totalTickets, len(reserveTickets))
}

// func Test_AllocateTickets_NoRequestsDuringOpenWindow(t *testing.T) {
// 	// AutoMigrate models
// 	db.AutoMigrate(&models.TicketRelease{}, &models.TicketRequest{}, &models.Ticket{})

// 	// Create AllocateTicketsService
// 	ats := services.NewAllocateTicketsService(db)

// 	var totalTickets int = 100

// 	// Create a TicketRelease with OpenWindowDuration
// 	tr := models.TicketRelease{
// 		Open: uint(time.Now().Unix() - 1000),
// 		TicketReleaseMethodDetail: models.TicketReleaseMethodDetail{
// 			OpenWindowDuration: 30, // 30 seconds window
// 			TicketReleaseMethod: models.TicketReleaseMethod{
// 				MethodName: string(models.FCFS_LOTTERY),
// 			},
// 		},
// 		TicketTypes: []models.TicketType{
// 			{QuantityTotal: uint(totalTickets)},
// 		},
// 	}

// 	requests := 1000

// 	var ticketRequest models.TicketRequest
// 	for i := 0; i < requests; i++ {
// 		ticketRequest = models.TicketRequest{
// 			TicketReleaseID: tr.ID,
// 			Model: gorm.Model{
// 				CreatedAt: time.Now().Add(time.Duration(100) * time.Second),
// 			},
// 		}

// 		db.Create(&ticketRequest)
// 	}

// 	// Allocate tickets
// 	err := ats.AllocateTickets(tr)

// 	// Validate
// 	assert.Nil(t, err)

// 	// We should have 100 tickets allocated
// 	var tickets []models.Ticket
// 	var reserveTickets []models.Ticket
// 	db.Where("is_reserve = ?", false).Find(&tickets)
// 	db.Where("is_reserve = ?", true).Find(&reserveTickets)

// 	// Filter

// 	assert.Equal(t, totalTickets, len(tickets))
// 	assert.Equal(t, requests-totalTickets, len(reserveTickets))
// }
