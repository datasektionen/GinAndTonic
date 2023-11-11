package services

import (
	"errors"
	"math/rand"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

type AllocateTicketsService struct {
	DB *gorm.DB
}

func NewAllocateTicketsService(db *gorm.DB) *AllocateTicketsService {
	return &AllocateTicketsService{DB: db}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (ats *AllocateTicketsService) AllocateTickets(ticketRelease models.TicketRelease) error {
	method := ticketRelease.TicketReleaseMethodDetail.TicketReleaseMethod

	if method.MethodName == "" {
		// Raise error
		return errors.New("No method name specified")
	}

	switch method.MethodName {
	case string(models.FCFS_LOTTERY):
		return ats.allocateFCFSLotteryTickets(ticketRelease)
	default:
		return nil
	}
}

func (ats *AllocateTicketsService) allocateFCFSLotteryTickets(ticketRelease models.TicketRelease) error {
	methodDetail := ticketRelease.TicketReleaseMethodDetail

	// Calculate the deadline for eligible requests
	deadline := utils.ConvertUNIXTimeToDateTime(int64(ticketRelease.Open + methodDetail.OpenWindowDuration))

	// Fetch all ticket requests directly from the database
	var allTicketRequests []models.TicketRequest
	if err := ats.DB.Where("ticket_release_id = ? AND is_handled = ?", ticketRelease.ID, false).Find(&allTicketRequests).Error; err != nil {
		return err
	}

	eligibleTicketRequestsForLottery := make([]models.TicketRequest, 0)
	notEligibleTicketRequests := make([]models.TicketRequest, 0)

	// Split ticket requests based on eligibility
	for _, tr := range allTicketRequests {
		if tr.CreatedAt.Before(deadline) || tr.CreatedAt.Equal(deadline) {
			eligibleTicketRequestsForLottery = append(eligibleTicketRequestsForLottery, tr)
		} else {
			notEligibleTicketRequests = append(notEligibleTicketRequests, tr)
		}
	}

	// Fetch total available tickets directly
	var availableTickets int
	for _, ticketType := range ticketRelease.TicketTypes {
		availableTickets += int(ticketType.QuantityTotal)
	}

	if len(eligibleTicketRequestsForLottery) > availableTickets {
		rand.Shuffle(len(eligibleTicketRequestsForLottery), func(i, j int) {
			eligibleTicketRequestsForLottery[i], eligibleTicketRequestsForLottery[j] = eligibleTicketRequestsForLottery[j], eligibleTicketRequestsForLottery[i]
		})

		for i := 0; i < len(eligibleTicketRequestsForLottery); i++ {
			if i < availableTickets {
				if err := ats.allocateTicket(eligibleTicketRequestsForLottery[i]); err != nil {
					return err
				}
			} else {
				if err := ats.allocateReserveTicket(eligibleTicketRequestsForLottery[i]); err != nil {
					return err
				}
			}
		}
	} else {
		for _, ticketRequest := range eligibleTicketRequestsForLottery {
			if err := ats.allocateTicket(ticketRequest); err != nil {
				return err
			}
		}
	}

	remainingTickets := availableTickets - len(eligibleTicketRequestsForLottery)
	for _, ticketRequest := range notEligibleTicketRequests {
		if remainingTickets > 0 {
			if err := ats.allocateTicket(ticketRequest); err != nil {
				return err
			}
			remainingTickets--
		} else {
			if err := ats.allocateReserveTicket(ticketRequest); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ats *AllocateTicketsService) allocateTicket(ticketRequest models.TicketRequest) error {
	ticketRequest.IsHandled = true
	if err := ats.DB.Save(&ticketRequest).Error; err != nil {
		return err
	}

	ticket := models.Ticket{
		TicketRequestID: ticketRequest.ID,
		IsPaid:          false,
		IsReserve:       false,
		UserUGKthID:     ticketRequest.UserUGKthID,
	}

	if err := ats.DB.Create(&ticket).Error; err != nil {
		return err
	}

	return nil
}

func (ats *AllocateTicketsService) allocateReserveTicket(ticketRequest models.TicketRequest) error {
	ticketRequest.IsHandled = true
	if err := ats.DB.Save(&ticketRequest).Error; err != nil {
		return err
	}
	ticket := models.Ticket{
		TicketRequestID: ticketRequest.ID,
		IsPaid:          false,
		IsReserve:       true,
		UserUGKthID:     ticketRequest.UserUGKthID,
	}

	if err := ats.DB.Create(&ticket).Error; err != nil {
		return err
	}

	return nil
}
