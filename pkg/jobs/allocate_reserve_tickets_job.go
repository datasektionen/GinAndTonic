package jobs

import (
	"errors"
	"os"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var allocator_logger = logrus.New()

func init() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	if _, err := os.Stat("logs/allocate_reserve_tickets_job.log"); os.IsNotExist(err) {
		os.Create("logs/allocate_reserve_tickets_job.log")
	}

	allocator_log_file, err := os.OpenFile("logs/allocate_reserve_tickets_job.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		allocator_logger.Fatal(err)
	}

	// Set log output to the file
	allocator_logger.SetOutput(allocator_log_file)

	// Set log level
	allocator_logger.SetLevel(logrus.InfoLevel)

	// Log as JSON for structured logging
	allocator_logger.SetFormatter(&logrus.JSONFormatter{})
}

func AllocateReserveTicketsJob(db *gorm.DB) error {
	start := time.Now()

	closedTicketReleases, err := models.GetClosedTicketReleases(db)
	if err != nil {
		return err
	}

	if len(closedTicketReleases) == 0 {
		allocator_logger.Info("No closed ticket releases")
		return nil
	}

	allocator_logger.WithFields(logrus.Fields{
		"number_of_closed_ticket_releases": len(closedTicketReleases),
	}).Info("Starting to process closed ticket releases")

	for _, ticketRelease := range closedTicketReleases {
		err := process_mpartj(db, ticketRelease)
		if err != nil {
			// Depending on your error handling strategy, you can decide whether
			// to continue processing the next ticket release or return the error.
			// For now, we're just printing the error and moving to the next ticket release.
		}
	}

	elapsed := time.Since(start)
	allocator_logger.Infof("AllocateReserveTicketsJob took %s", elapsed)

	return nil
}

func ManuallyProcessAllocateReserveTicketsJob(db *gorm.DB, ticketReleaseID uint) error {
	start := time.Now()

	var ticketRelease *models.TicketRelease
	err := db.First(&ticketRelease, ticketReleaseID).Error

	if err != nil {
		return err
	}

	if ticketRelease == nil {
		allocator_logger.WithFields(logrus.Fields{
			"ticket_release_id": ticketReleaseID,
		}).Info("Ticket release does not exist")
		return nil
	}

	allocator_logger.WithFields(logrus.Fields{
		"ticket_release_id": ticketReleaseID,
	}).Info("Starting to process ticket release")

	err = process_mpartj(db, *ticketRelease)
	if err != nil {
		return err
	}

	elapsed := time.Since(start)
	allocator_logger.Infof("ManuallyProcessAllocateReserveTicketsJob took %s", elapsed)

	return nil
}

func process_mpartj(db *gorm.DB, ticketRelease models.TicketRelease) error {
	// Start by getting all ticket releases that have not allocated tickets

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		allocator_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error beginning transaction: %s", tx.Error.Error())

		return tx.Error
	}

	// This case should never happen since if tickets are allocated then the ticket release is closed
	// but we check anyway
	if !ticketRelease.HasAllocatedTickets {
		allocator_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Info("Ticket release has not allocated tickets")
		tx.Rollback()
		return nil
	}

	if ticketRelease.HasPromoCode() {
		allocator_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Info("Ticket release has a promo code")
		tx.Rollback()
		return nil
	}

	// availableTickets := ticketRelease.TicketsAvailable
	payWithin := ticketRelease.PayWithin
	// Calculate when ticket should have been paid depending on the when the ticket was updated

	// Get all allocated tickets that isnt a reserve ticket or deleted
	allocatedTickets, err := models.GetAllTicketsToTicketRelease(tx, ticketRelease.ID)
	if err != nil {
		allocator_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error getting all tickets to ticket release with ID %d: %s", ticketRelease.ID, err.Error())
		tx.Rollback()
		return err
	}

	// Get all reserved tickets that isnt deleted
	reservedTickets, err := models.GetAllReserveTicketsToTicketRelease(tx, ticketRelease.ID)
	if err != nil {
		allocator_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error getting all reserve tickets to ticket release with ID %d: %s", ticketRelease.ID, err.Error())
		tx.Rollback()
		return err
	}

	allocator_logger.WithFields(logrus.Fields{
		"id":                          ticketRelease.ID,
		"number_of_allocated_tickets": len(allocatedTickets),
		"number_of_reserved_tickets":  len(reservedTickets),
	}).Info("Got all allocated and reserved tickets")

	// Initialize newReserveTickets by taking the tickets available and subtracting the number of allocated tickets that hasnt been deleted
	var newReserveTickets int64 = int64(ticketRelease.TicketsAvailable) - int64(len(allocatedTickets))

	var newlyAllocatedTicketIDs []int
	var newlyRemovedTicket []*models.Ticket

	for _, ticket := range allocatedTickets {
		// We start by iterating through all allocated tickets
		// Here we want to check if the ticket has been paid or if the ticket has not been paid within the time limit.
		var mustPayBefore time.Time

		if payWithin != nil {
			mustPayBefore = utils.MustPayBefore(int(*payWithin), ticket.UpdatedAt)
		} else {
			allocator_logger.WithFields(logrus.Fields{
				"id": ticketRelease.ID,
			}).Error("Pay within is nil")
			tx.Rollback()

			return errors.New("Pay within is nil")
		}

		if ticket.IsPaid || time.Now().Before(mustPayBefore) {
			// If the ticket has been paid or we are still within the time limit then we continue
			continue
		}

		// If we reach this point then the ticket has not been paid within the time limit
		// We can say that the user looses the ticket, we set the ticket to deleted and
		// increment the number of new tickets to be allocated from the reserve list
		err := ticket.Delete(tx)
		if err != nil {
			allocator_logger.WithFields(logrus.Fields{
				"id": ticketRelease.ID,
			}).Errorf("Error deleting ticket with ID %v: %s", ticketRelease.ID, err.Error())

			continue
		}

		allocator_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Infof("Deleted ticket with ID %d", ticket.ID)

		newlyRemovedTicket = append(newlyRemovedTicket, &ticket)
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
					allocator_logger.WithFields(logrus.Fields{
						"id": ticketRelease.ID,
					}).Errorf("Error saving ticket with ID %d: %s", ticketRelease.ID, err.Error())
				}

				continue
			}

			// We want to allocate the ticket
			// We set the ticket to not be a reserve ticket and set the reserve number to 0
			var isPaid bool = false
			if ticket.TicketRequest.TicketType.Price == 0 {
				isPaid = true
			}

			ticket.IsReserve = false
			ticket.ReserveNumber = 0
			ticket.IsPaid = isPaid

			if err := tx.Save(&ticket).Error; err != nil {
				allocator_logger.WithFields(logrus.Fields{
					"id": ticketRelease.ID,
				}).Errorf("Error saving ticket with ID %d: %s", ticketRelease.ID, err.Error())

				continue
			}

			allocator_logger.WithFields(logrus.Fields{
				"id": ticketRelease.ID,
			}).Infof("Allocated ticket with ID %d", ticket.ID)

			newlyAllocatedTicketIDs = append(newlyAllocatedTicketIDs, int(ticket.ID))
		}

	}

	err = tx.Commit().Error

	if err != nil {
		allocator_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error committing transaction: %s", err.Error())

		return err
	}

	for _, ticketID := range newlyAllocatedTicketIDs {
		err := Notify_ReserveTicketConvertedAllocation(db, ticketID)

		if err != nil {
			allocator_logger.WithFields(logrus.Fields{
				"id": ticketRelease.ID,
			}).Errorf("Error notifying user about ticket allocation: %s", err.Error())
		}
	}

	for _, ticket := range newlyRemovedTicket {
		err := Notify_TicketNotPaidInTime(db, ticket)

		if err != nil {
			allocator_logger.WithFields(logrus.Fields{
				"id": ticketRelease.ID,
			}).Errorf("Error notifying user about ticket not being paid in time: %s", err.Error())
		}
	}

	allocator_logger.WithFields(logrus.Fields{
		"id": ticketRelease.ID,
	}).Info("Finished processing ticket release")

	return nil

}
