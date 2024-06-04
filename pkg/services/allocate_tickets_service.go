package services

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	allocate_fcfs "github.com/DowLucas/gin-ticket-release/pkg/services/allocate_fcfc"
	"github.com/DowLucas/gin-ticket-release/pkg/services/allocate_service"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
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

func (ats *AllocateTicketsService) AllocateTickets(ticketRelease *models.TicketRelease, allocateTicketsRequest *types.AllocateTicketsRequest) error {
	method := ticketRelease.TicketReleaseMethodDetail.TicketReleaseMethod
	var tickets []*models.Ticket
	var err error

	if method.MethodName == "" {
		// Raise error
		return errors.New("no method name specified")
	}

	// Check if allocation has already been done
	if ticketRelease.HasAllocatedTickets {
		return errors.New("tickets already allocated")
	}

	// Before allocating tickets, check if the ticket release is open
	// If it is open then we close it
	// We use transaction to ensure that the ticket release is closed

	tx := ats.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(ticketRelease).Update("has_allocated_tickets", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	switch method.MethodName {
	case string(models.FCFS_LOTTERY):
		tickets, err = ats.allocateFCFSLotteryTickets(ticketRelease, tx)

		if err != nil {
			tx.Rollback()
			return err
		}

		if len(tickets) > 0 {
			// Notify the users that the tickets have been allocated
			for _, ticket := range tickets {
				var err error
				if !ticket.IsReserve {
					err = Notify_TicketAllocationCreated(tx,
						int(ticket.ID),
						&ticketRelease.PaymentDeadline.OriginalDeadline)
				} else {
					err = Notify_ReserveTicketAllocationCreated(tx, int(ticket.ID))
				}

				if err != nil {
					tx.Rollback()
					fmt.Println(err)
					return err
				}
			}
		}
	case string(models.RESERVED_TICKET_RELEASE):
		tickets, err = ats.allocateReservedTickets(ticketRelease, tx)

		if err != nil {
			tx.Rollback()
			return err
		}

		if len(tickets) > 0 {
			// Notify the users that the tickets have been allocated
			for _, ticket := range tickets {
				var err error
				if !ticket.IsReserve {
					err = Notify_TicketAllocationCreated(tx, int(ticket.ID), nil)
				} else {
					err = Notify_ReserveTicketAllocationCreated(tx, int(ticket.ID))
				}

				if err != nil {
					tx.Rollback()
					fmt.Println(err)
					return err
				}
			}
		}
	case string(models.FCFS):
		tickets, err := allocate_fcfs.AllocateFCFSTickets(ticketRelease, tx)

		if err != nil {
			fmt.Println(err)
			tx.Rollback()
			return err
		}

		if len(tickets) > 0 {
			// Notify the users that the tickets have been allocated
			for _, ticket := range tickets {
				var err error
				if !ticket.IsReserve {
					err = Notify_TicketAllocationCreated(tx, int(ticket.ID), &ticketRelease.PaymentDeadline.OriginalDeadline)
				} else {
					err = Notify_ReserveTicketAllocationCreated(tx, int(ticket.ID))
				}

				if err != nil {
					tx.Rollback()
					fmt.Println(err)
					return err
				}
			}
		}

	default:
		tx.Rollback()
		return errors.New("unknown ticket release method")
	}

	return tx.Commit().Error
}

func (ats *AllocateTicketsService) allocateFCFSLotteryTickets(
	ticketRelease *models.TicketRelease,
	tx *gorm.DB) (allTickets []*models.Ticket, err error) {
	var reserveNumber uint

	methodDetail := ticketRelease.TicketReleaseMethodDetail

	// Calculate the deadline for eligible requests
	deadline := utils.ConvertUNIXTimeToDateTime(ticketRelease.Open.Add(time.Duration(methodDetail.OpenWindowDuration) * time.Minute).Unix())

	// Fetch all ticket requests directly from the database
	allTicketOrders, err := models.GetAllValidTicketOrdersToTicketRelease(tx, ticketRelease.ID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(allTicketOrders) == 0 {
		// return empty array if there are no ticket requests
		return allTickets, nil
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
				ticket, err := allocate_service.AllocateTicket(eligibleTicketRequestsForLottery[i], tx)
				if err != nil {
					return nil, err
				}
				allTickets = append(allTickets, ticket)
			} else {
				ticket, err := allocate_service.AllocateReserveTicket(eligibleTicketRequestsForLottery[i], reserveNumber, tx)
				if err != nil {
					return nil, err
				}
				reserveNumber++
				allTickets = append(allTickets, ticket)
			}
		}
	} else {
		for _, ticketRequest := range eligibleTicketRequestsForLottery {
			ticket, err := allocate_service.AllocateTicket(ticketRequest, tx)
			if err != nil {
				return nil, err
			}
			allTickets = append(allTickets, ticket)
		}
	}

	remainingTickets := availableTickets - len(eligibleTicketRequestsForLottery)
	reserveNumber = 1
	for _, ticketRequest := range notEligibleTicketRequests {
		if remainingTickets > 0 {
			ticket, err := allocate_service.AllocateTicket(ticketRequest, tx)
			if err != nil {
				return nil, err
			}
			allTickets = append(allTickets, ticket)
			remainingTickets--
		} else {
			ticket, err := allocate_service.AllocateReserveTicket(ticketRequest, reserveNumber, tx)
			if err != nil {
				return nil, err
			}
			allTickets = append(allTickets, ticket)
			reserveNumber++
		}
	}

	return allTickets, nil
}

func (ats *AllocateTicketsService) allocateReservedTickets(ticketRelease *models.TicketRelease, tx *gorm.DB) (tickets []*models.Ticket, err error) {
	// Fetch all ticket requests directly from the database
	var reserveNumber uint = 1
	var allTicketRequests []models.TicketRequest
	if err := tx.Preload("TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Where("ticket_release_id = ? AND is_handled = ?", ticketRelease.ID, false).Find(&allTicketRequests).Order("created_at").Error; err != nil {
		return nil, err
	}

	// Fetch total available tickets directly
	var availableTickets int = ticketRelease.TicketsAvailable

	// Give all users tickets up to the available tickets, give the rest reserve tickets
	for i, ticketRequest := range allTicketRequests {
		if i < availableTickets {
			ticket, err := allocate_service.AllocateTicket(ticketRequest, tx)
			if err != nil {
				return nil, err
			}
			tickets = append(tickets, ticket)
		} else {
			ticket, err := allocate_service.AllocateReserveTicket(ticketRequest, reserveNumber, tx)
			if err != nil {
				return nil, err
			}
			tickets = append(tickets, ticket)
			reserveNumber++
		}
	}

	return tickets, nil
}

func SelectivelyAllocateTicketRequest(db *gorm.DB, ticketRequestID int) error {
	// Use your database layer to find the ticket request by ID and allocate it
	// This is just a placeholder implementation, replace it with your actual code
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ticketRequest models.TicketRequest
	err := tx.Preload("User").
		Preload("TicketRelease.TicketReleaseMethodDetail.TicketReleaseMethod").
		Preload("TicketRelease.PaymentDeadline").
		Where("id = ?", ticketRequestID).First(&ticketRequest).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// Check if ticket request is already handled
	// If the ticket request is already handled, it cannot be allocated
	if ticketRequest.IsHandled {
		tx.Rollback()
		return errors.New("ticket request is already handled")
	}

	// Alocate the ticket
	ticket, err := allocate_service.AllocateTicket(ticketRequest, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = Notify_TicketAllocationCreated(tx, int(ticket.ID), nil)

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	return nil
}

func SelectivelyAllocateTicketRequests(db *gorm.DB, ticketRequestIDs []int) *types.ErrorResponse {
	// Use your database layer to find the ticket requests by ID and allocate them
	// This is just a placeholder implementation, replace it with your actual code

	for _, ticketRequestID := range ticketRequestIDs {
		err := SelectivelyAllocateTicketRequest(db, ticketRequestID)
		if err != nil {
			return &types.ErrorResponse{StatusCode: 400, Message: err.Error()}
		}
	}

	return nil
}
