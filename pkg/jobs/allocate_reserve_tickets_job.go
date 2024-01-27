package jobs

import (
	"fmt"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func AllocateReserveTicketsJob(db *gorm.DB) error {
	closedTicketReleases, err := models.GetClosedTicketReleases(db)
	if err != nil {
		return err
	}

	for _, ticketRelease := range closedTicketReleases {
		err := process(db, ticketRelease)
		if err != nil {
			// Depending on your error handling strategy, you can decide whether
			// to continue processing the next ticket release or return the error.
			// For now, we're just printing the error and moving to the next ticket release.
			fmt.Printf("error processing ticket release %d: %v\n", ticketRelease.ID, err)
		}
	}

	return nil
}

func process(db *gorm.DB, ticketRelease models.TicketRelease) error {
	// Start by getting all ticket releases that have not allocated tickets

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// This case should never happen since if tickets are allocated then the ticket release is closed
	// but we check anyway
	if !ticketRelease.HasAllocatedTickets {
		fmt.Println("Ticket release has not allocated tickets")
	}

	// If the ticket release has a promo code then we skip it
	if ticketRelease.PromoCode != nil {
		fmt.Println("Ticket release has a promo code")
	}

	// availableTickets := ticketRelease.TicketsAvailable
	payWithin := ticketRelease.PayWithin
	// Calculate when ticket should have been paid depending on the when the ticket was updated

	// Get all allocated tickets that isnt a reserve ticket or deleted
	allocatedTickets, err := models.GetAllTicketsToTicketRelease(tx, ticketRelease.ID)
	if err != nil {
		fmt.Println(err)
	}

	// Get all reserved tickets that isnt deleted
	reservedTickets, err := models.GetAllReserveTicketsToTicketRelease(tx, ticketRelease.ID)
	if err != nil {
		fmt.Println(err)
	}

	// Initialize newReserveTickets by taking the tickets available and subtracting the number of allocated tickets that hasnt been deleted
	var newReserveTickets int64 = int64(ticketRelease.TicketsAvailable) - int64(len(allocatedTickets))

	for _, ticket := range allocatedTickets {
		// We start by iterating through all allocated tickets
		// Here we want to check if the ticket has been paid or if the ticket has not been paid within the time limit.
		var mustPayBefore time.Time

		if payWithin != nil {
			mustPayBefore = ticket.UpdatedAt.Add(time.Duration(*payWithin) * time.Hour)
		} else {
			tx.Rollback()
			return fmt.Errorf("Pay within is nil")
		}

		if ticket.IsPaid || time.Now().Before(mustPayBefore) {
			// If the ticket has been paid or we are still within the time limit then we continue
			continue
		}

		// If we reach this point then the ticket has not been paid within the time limit
		// We can say that the user looses the ticket, we set the ticket to deleted and
		// increment the number of new tickets to be allocated from the reserve list
		if err := ticket.Delete(tx); err != nil {
			fmt.Println(err)
			continue
		}
		// TODO: Notify user that ticket has been deleted

		// Increment number of new tickets to be allocated from the reserve list
		newReserveTickets++
	}

	// We now have the number of new tickets to be allocated from the reserve list
	if newReserveTickets > 0 {
		// We now want to allocate newReserveTickets from the reserve list
		// Each ticket has a reserve number, we want to allocate the tickets with the lowest reserve number
		// The list is already sorted by reserve number so we can just iterate through the list
		var ticket models.Ticket
		for i := 0; i < len(reservedTickets); i++ {
			ticket = reservedTickets[i]
			// If i is greater than or equal to newReserveTickets then we have allocated all tickets
			// And we want to update the reserve number for the remaining tickets
			// This should be i - newReserveTickets + 1
			if i >= int(newReserveTickets) {
				ticket.ReserveNumber = uint(i - int(newReserveTickets) + 1)

				if err := tx.Save(&ticket).Error; err != nil {
					fmt.Println(err)
				}

				continue
			}

			// We want to allocate the ticket
			// We set the ticket to not be a reserve ticket and set the reserve number to 0
			ticket.IsReserve = false
			ticket.ReserveNumber = 0

			if err := tx.Save(&ticket).Error; err != nil {
				fmt.Println(err)
				continue
			}

			// TODO Notify user that ticket has been allocated

		}

	}

	err = tx.Commit().Error
	if err != nil {
		fmt.Println(err)
	}

	return nil

}
