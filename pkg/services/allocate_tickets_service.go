package services

import (
	"errors"
	"math/rand"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

var r *rand.Rand

type AllocateTicketsService struct {
	DB *gorm.DB
}

func NewAllocateTicketsService(db *gorm.DB) *AllocateTicketsService {
	return &AllocateTicketsService{DB: db}
}

func init() {
	seed := time.Now().UnixNano()
	r = rand.New(rand.NewSource(seed))
}

func (ats *AllocateTicketsService) AllocateTickets(ticketRelease *models.TicketRelease) error {
	method := ticketRelease.TicketReleaseMethodDetail.TicketReleaseMethod

	if method.MethodName == "" {
		// Raise error
		return errors.New("No method name specified")
	}

	// Check if allocation has already been done
	if ticketRelease.HasAllocatedTickets {
		return errors.New("Tickets already allocated")
	}

	// Before allocating tickets, check if the ticket release is open
	// If it is open then we close it
	// We use transaction to ensure that the ticket release is closed

	tx := ats.DB.Begin()

	if err := tx.Model(ticketRelease).Update("has_allocated_tickets", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	switch method.MethodName {
	case string(models.FCFS_LOTTERY):
		err := ats.allocateFCFSLotteryTickets(ticketRelease, tx)

		if err != nil {
			tx.Rollback()
			return err
		}
		break
	case string(models.RESERVED_TICKET_RELEASE):
		err := ats.allocateReservedTickets(ticketRelease, tx)

		if err != nil {
			tx.Rollback()
			return err
		}
		break

	default:
		tx.Rollback()
		return errors.New("Unknown ticket release method")
	}

	tx.Commit()

	return nil
}

func (ats *AllocateTicketsService) allocateFCFSLotteryTickets(ticketRelease *models.TicketRelease, tx *gorm.DB) error {
	var reserveNumber uint
	methodDetail := ticketRelease.TicketReleaseMethodDetail

	// Calculate the deadline for eligible requests
	deadline := utils.ConvertUNIXTimeToDateTime(int64(ticketRelease.Open + methodDetail.OpenWindowDuration))

	// Fetch all ticket requests directly from the database
	allTicketRequests, err := models.GetAllValidTicketRequestsToTicketRelease(tx, ticketRelease.ID)

	if err != nil {
		tx.Rollback()
		return err
	}

	if len(allTicketRequests) == 0 {
		return errors.New("No ticket requests to allocate")
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
	var availableTickets int = ticketRelease.TicketsAvailable

	if len(eligibleTicketRequestsForLottery) > availableTickets {
		rand.Shuffle(len(eligibleTicketRequestsForLottery), func(i, j int) {
			eligibleTicketRequestsForLottery[i], eligibleTicketRequestsForLottery[j] = eligibleTicketRequestsForLottery[j], eligibleTicketRequestsForLottery[i]
		})
		reserveNumber = 1
		for i := 0; i < len(eligibleTicketRequestsForLottery); i++ {
			if i < availableTickets {
				if err := ats.AllocateTicket(eligibleTicketRequestsForLottery[i], tx); err != nil {
					return err
				}
			} else {
				if err := ats.AllocateReserveTicket(eligibleTicketRequestsForLottery[i], reserveNumber, tx); err != nil {
					return err
				}
				reserveNumber++
			}
		}
	} else {
		for _, ticketRequest := range eligibleTicketRequestsForLottery {
			if err := ats.AllocateTicket(ticketRequest, tx); err != nil {
				return err
			}
		}
	}

	remainingTickets := availableTickets - len(eligibleTicketRequestsForLottery)
	reserveNumber = 1
	for _, ticketRequest := range notEligibleTicketRequests {
		if remainingTickets > 0 {
			if err := ats.AllocateTicket(ticketRequest, tx); err != nil {
				return err
			}
			remainingTickets--
		} else {
			if err := ats.AllocateReserveTicket(ticketRequest, reserveNumber, tx); err != nil {
				return err
			}
			reserveNumber++
		}
	}

	return nil
}

func (ats *AllocateTicketsService) allocateReservedTickets(ticketRelease *models.TicketRelease, tx *gorm.DB) error {
	// Fetch all ticket requests directly from the database
	var reserveNumber uint = 1
	var allTicketRequests []models.TicketRequest
	if err := tx.Where("ticket_release_id = ? AND is_handled = ?", ticketRelease.ID, false).Find(&allTicketRequests).Order("created_at").Error; err != nil {
		return err
	}

	// Fetch total available tickets directly
	var availableTickets int = ticketRelease.TicketsAvailable

	// Give all users ticekts up to the available tickets, give the rest reserve tickets
	for i, ticketRequest := range allTicketRequests {
		if i < availableTickets {
			if err := ats.AllocateTicket(ticketRequest, tx); err != nil {
				return err
			}
		} else {
			if err := ats.AllocateReserveTicket(ticketRequest, reserveNumber, tx); err != nil {
				return err
			}
			reserveNumber++
		}
	}

	return nil
}

func (ats *AllocateTicketsService) AllocateTicket(ticketRequest models.TicketRequest, tx *gorm.DB) error {
	ticketRequest.IsHandled = true
	if err := tx.Save(&ticketRequest).Error; err != nil {
		return err
	}

	ticket := models.Ticket{
		TicketRequestID: ticketRequest.ID,
		IsReserve:       false,
		UserUGKthID:     ticketRequest.UserUGKthID,
	}

	if err := tx.Create(&ticket).Error; err != nil {
		return err
	}

	return nil
}

func (ats *AllocateTicketsService) AllocateReserveTicket(
	ticketRequest models.TicketRequest,
	reserveNumber uint,
	tx *gorm.DB) error {
	ticketRequest.IsHandled = true
	if err := tx.Save(&ticketRequest).Error; err != nil {
		return err
	}

	ticket := models.Ticket{
		TicketRequestID: ticketRequest.ID,
		ReserveNumber:   reserveNumber,
		IsReserve:       true,
		UserUGKthID:     ticketRequest.UserUGKthID,
	}

	if err := tx.Create(&ticket).Error; err != nil {
		return err
	}

	return nil
}
