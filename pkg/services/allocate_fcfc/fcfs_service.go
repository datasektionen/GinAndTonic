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
	allTicketRequests, err := models.GetAllValidTicketRequestToTicketReleaseOrderedByCreatedAt(tx, ticketRelease.ID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(allTicketRequests) == 0 {
		// return empty list of tickets
		return []*models.Ticket{}, nil
	}

	var numberOfTicketsAllocated int = 0
	var tickets []*models.Ticket

	for _, ticketRequest := range allTicketRequests {
		// Check if the ticket request is handled
		if ticketRequest.IsHandled {
			continue
		}

		var ticket *models.Ticket
		if numberOfTicketsAllocated >= ticketRelease.TicketsAvailable {
			ticket, err = allocate_service.AllocateReserveTicket(ticketRequest,
				uint(numberOfTicketsAllocated-ticketRelease.TicketsAvailable),
				tx)
		} else {
			ticket, err = allocate_service.AllocateTicket(ticketRequest, tx)
		}

		if err != nil {
			tx.Rollback()
			return nil, err
		}

		numberOfTicketsAllocated++

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}
