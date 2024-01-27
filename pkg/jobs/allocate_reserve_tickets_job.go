package jobs

import (
	"fmt"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func AllocateReserveTicketsJob(db *gorm.DB) {
	// Start by getting all ticket releases that have not allocated tickets
	openTicketReleases, err := models.GetOpenTicketReleases(db)

	if err != nil {
		return
	}

	for _, ticketRelease := range openTicketReleases {
		// This case should never happen since if tickets are allocated then the ticket release is closed
		// but we check anyway
		if ticketRelease.HasAllocatedTickets {
			continue
		}

		// If the ticket release has a promo code then we skip it
		if ticketRelease.PromoCode != nil {
			continue
		}

		// availableTickets := ticketRelease.TicketsAvailable
		payWithin := ticketRelease.PayWithin
		// Calculate when ticket should have been paid depending on the when the ticket was updated

		// Get all allocated tickets that isnt a reserve ticket or deleted
		allocatedTickets, err := models.GetAllTicketsToTicketRelease(db, ticketRelease.ID)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Get all reserved tickets that isnt deleted
		reservedTickets, err := models.GetAllReserveTicketsToTicketRelease(db, ticketRelease.ID)
		if err != nil {
			fmt.Println(err)
			continue
		}

		var newReserveTickets int64 = 0

		for _, ticket := range allocatedTickets {
			// We start by iterating through all allocated tickets
			// Here we want to check if the ticket has been paid or if the ticket has not been paid within the time limit.
			mustPayBefore := ticket.UpdatedAt.Add(time.Duration(*payWithin) * time.Hour)

			if ticket.IsPaid || time.Now().Before(mustPayBefore) {
				// If the ticket has been paid or we are still within the time limit then we continue
				continue
			}

			// If we reach this point then the ticket has not been paid within the time limit
			// We can say that the user looses the ticket, we set the ticket to deleted and
			// increment the number of new tickets to be allocated from the reserve list
			if err := ticket.Delete(db); err != nil {
				fmt.Println(err)
				continue
			}
			// TODO: Notify user that ticket has been deleted

			// Increment number of new tickets to be allocated from the reserve list
			newReserveTickets++

			println("Ticket: ", ticket.ID)
		}

		// We now have the number of new tickets to be allocated from the reserve list
		if newReserveTickets > 0 {
			// We now want to allocate newReserveTickets from the reserve list
			// Each ticket has a reserve number, we want to allocate the tickets with the lowest reserve number
			// The list is already sorted by reserve number so we can just iterate through the list

			var ticket models.Ticket
			for i := 0; i < len(reservedTickets); i++ {
				// If i is greater than or equal to newReserveTickets then we have allocated all tickets
				// And we want to update the reserve number for the remaining tickets
				// This should be i - newReserveTickets + 1
				if i >= int(newReserveTickets) {
					ticket.ReserveNumber = uint(i - int(newReserveTickets) + 1)

					if err := db.Save(&ticket).Error; err != nil {
						fmt.Println(err)
					}

					continue
				}

				// We want to allocate the ticket
				// We set the ticket to not be a reserve ticket and set the reserve number to 0
				ticket.IsReserve = false
				ticket.ReserveNumber = 0

				if err := db.Save(&ticket).Error; err != nil {
					fmt.Println(err)
					continue
				}

				// TODO Notify user that ticket has been allocated

			}

		}

		println("Allocated tickets: ", len(allocatedTickets))
		println("Reserved tickets: ", len(reservedTickets))
	}

}
