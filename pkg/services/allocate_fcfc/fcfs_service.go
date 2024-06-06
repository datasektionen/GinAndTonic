package allocate_fcfs

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services/allocate_service"
	"gorm.io/gorm"
)

func AllocateFCFSTickets(ticketRelease *models.TicketRelease, tx *gorm.DB) ([]*models.Ticket, error) {
	// Allocate tickets in a first-come-first-serve manner
	// The first ticket request to come in will be the first to be allocated a ticket
	// The last ticket request to come in will be the last to be allocated a ticket
	// Fetch all ticket requests directly from the database
	allTicketOrders, err := models.GetAllValidTicketOrdersToTicketReleaseOrderedByCreatedAt(tx, ticketRelease.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(allTicketOrders) == 0 {
		// return empty list of tickets
		return []*models.Ticket{}, nil
	}

	var numberOfTicketsAllocated int = 0
	var tickets []*models.Ticket

	for _, ticketOrder := range allTicketOrders {
		if ticketOrder.IsHandled {
			continue
		}

		for _, ticket := range ticketOrder.Tickets {
			// Check if the ticket request is handled

			if numberOfTicketsAllocated >= ticketRelease.TicketsAvailable {
				err = allocate_service.AllocateReserveTicket(&ticket, uint(numberOfTicketsAllocated-ticketRelease.TicketsAvailable), tx)
			} else {
				err = allocate_service.AllocateTicket(&ticket, ticketRelease.PaymentDeadline, tx)
			}

			if err != nil {
				tx.Rollback()
				return nil, err
			}

			numberOfTicketsAllocated++

			tickets = append(tickets, &ticket)
		}
	}

	return tickets, nil
}
