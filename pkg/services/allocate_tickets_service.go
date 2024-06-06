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
	unhandledTickets, err := models.GetAllUnhandledTicketsByTicketReleaseID(tx, ticketRelease.ID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(unhandledTickets) == 0 {
		// return empty array if there are no ticket requests
		return allTickets, nil
	}

	eligibleTicketOrdersForLottery := make([]models.Ticket, 0)
	notEligibleTicketOrders := make([]models.Ticket, 0)

	// Split ticket requests based on eligibility
	for _, tr := range unhandledTickets {
		if tr.CreatedAt.Before(deadline) || tr.CreatedAt.Equal(deadline) {
			eligibleTicketOrdersForLottery = append(eligibleTicketOrdersForLottery, tr)
		} else {
			notEligibleTicketOrders = append(notEligibleTicketOrders, tr)
		}
	}

	// Fetch total available tickets directly
	var availableTickets int = ticketRelease.TicketsAvailable

	if len(eligibleTicketOrdersForLottery) > availableTickets {
		rand.Shuffle(len(eligibleTicketOrdersForLottery), func(i, j int) {
			eligibleTicketOrdersForLottery[i], eligibleTicketOrdersForLottery[j] = eligibleTicketOrdersForLottery[j], eligibleTicketOrdersForLottery[i]
		})
		reserveNumber = 1
		for i := 0; i < len(eligibleTicketOrdersForLottery); i++ {
			if i < availableTickets {
				err := allocate_service.AllocateTicket(&eligibleTicketOrdersForLottery[i], ticketRelease.PaymentDeadline, tx)
				if err != nil {
					return nil, err
				}
				allTickets = append(allTickets, &eligibleTicketOrdersForLottery[i])
			} else {
				err := allocate_service.AllocateReserveTicket(&eligibleTicketOrdersForLottery[i], reserveNumber, tx)
				if err != nil {
					return nil, err
				}
				reserveNumber++
				allTickets = append(allTickets, &eligibleTicketOrdersForLottery[i])
			}
		}
	} else {
		for _, ticket := range eligibleTicketOrdersForLottery {
			err := allocate_service.AllocateTicket(&ticket, ticketRelease.PaymentDeadline, tx)
			if err != nil {
				return nil, err
			}
			allTickets = append(allTickets, &ticket)
		}
	}

	remainingTickets := availableTickets - len(eligibleTicketOrdersForLottery)
	reserveNumber = 1
	for _, ticket := range notEligibleTicketOrders {
		if remainingTickets > 0 {
			err := allocate_service.AllocateTicket(&ticket, ticketRelease.PaymentDeadline, tx)
			if err != nil {
				return nil, err
			}
			allTickets = append(allTickets, &ticket)
			remainingTickets--
		} else {
			err := allocate_service.AllocateReserveTicket(&ticket, reserveNumber, tx)
			if err != nil {
				return nil, err
			}
			allTickets = append(allTickets, &ticket)
			reserveNumber++
		}
	}

	return allTickets, nil
}

func (ats *AllocateTicketsService) allocateReservedTickets(ticketRelease *models.TicketRelease, tx *gorm.DB) (tickets []*models.Ticket, err error) {
	// Fetch all ticket requests directly from the database
	var reserveNumber uint = 1
	var allTickets []models.Ticket
	if err := tx.Preload("TicketType").
		Preload("TicketOrder.TicketRelease.Event").
		Preload("TicketOrder.TicketRelease.TicketReleaseMethodDetail").
		Where("ticket_release_id = ? AND is_handled = ?", ticketRelease.ID, false).Find(&allTickets).Order("created_at").Error; err != nil {
		return nil, err
	}

	// Fetch total available tickets directly
	var availableTickets int = ticketRelease.TicketsAvailable

	// Give all users tickets up to the available tickets, give the rest reserve tickets
	for i, ticket := range allTickets {
		if i < availableTickets {
			err := allocate_service.AllocateTicket(&ticket, ticketRelease.PaymentDeadline, tx)
			if err != nil {
				return nil, err
			}
			tickets = append(tickets, &ticket)
		} else {
			err := allocate_service.AllocateReserveTicket(&ticket, reserveNumber, tx)
			if err != nil {
				return nil, err
			}
			tickets = append(tickets, &ticket)
			reserveNumber++
		}
	}

	return tickets, nil
}

func SelectivelyAllocateTicketOrder(db *gorm.DB, ticketOrderId int) error {
	// Use your database layer to find the ticket request by ID and allocate it
	// This is just a placeholder implementation, replace it with your actual code
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ticketOrder models.TicketOrder
	err := tx.Preload("User").
		Preload("Tickets").
		Preload("TicketRelease.TicketReleaseMethodDetail.TicketReleaseMethod").
		Preload("TicketRelease.PaymentDeadline").
		Where("id = ?", ticketOrderId).First(&ticketOrder).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	if ticketOrder.IsHandled {
		tx.Rollback()
		return errors.New("ticket has already been handled")
	}

	// Check if ticket request is already handled
	// If the ticket request is already handled, it cannot be allocated
	for _, ticket := range ticketOrder.Tickets {
		// Alocate the ticket
		err = allocate_service.AllocateTicket(&ticket, ticket.TicketOrder.TicketRelease.PaymentDeadline, tx)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = Notify_TicketAllocationCreated(tx, int(ticket.ID), nil)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	return nil
}

func SelectivelyAllocateTicketOrders(db *gorm.DB, ticketOrderIds []int) *types.ErrorResponse {
	// Use your database layer to find the ticket requests by ID and allocate them
	// This is just a placeholder implementation, replace it with your actual code

	for _, ticketOrderId := range ticketOrderIds {
		err := SelectivelyAllocateTicketOrder(db, ticketOrderId)
		if err != nil {
			return &types.ErrorResponse{StatusCode: 400, Message: err.Error()}
		}
	}

	return nil
}
