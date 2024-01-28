package jobs

import (
	"errors"
	"os"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var log = logrus.New()

func init() {
	logFilePath := "./logs/allocate_reserve_tickets_job.log"

	// Create a file to write logs to. Append to existing file, create if not exists, writable.
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// Set log output to the file
	log.SetOutput(logFile)

	// Set log level
	log.SetLevel(logrus.InfoLevel)

	// Log as JSON for structured logging
	log.SetFormatter(&logrus.JSONFormatter{})
}

func AllocateReserveTicketsJob(db *gorm.DB) error {
	start := time.Now()

	closedTicketReleases, err := models.GetClosedTicketReleases(db)
	if err != nil {
		return err
	}

	if len(closedTicketReleases) == 0 {
		log.Info("No closed ticket releases")
		return nil
	}

	log.WithFields(logrus.Fields{
		"number_of_closed_ticket_releases": len(closedTicketReleases),
	}).Info("Starting to process closed ticket releases")

	for _, ticketRelease := range closedTicketReleases {
		err := process(db, ticketRelease)
		if err != nil {
			// Depending on your error handling strategy, you can decide whether
			// to continue processing the next ticket release or return the error.
			// For now, we're just printing the error and moving to the next ticket release.
		}
	}

	elapsed := time.Since(start)
	log.Infof("AllocateReserveTicketsJob took %s", elapsed)

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
		log.WithFields(logrus.Fields{
			"ticket_release_id": ticketReleaseID,
		}).Info("Ticket release does not exist")
		return nil
	}

	log.WithFields(logrus.Fields{
		"ticket_release_id": ticketReleaseID,
	}).Info("Starting to process ticket release")

	err = process(db, *ticketRelease)
	if err != nil {
		return err
	}

	elapsed := time.Since(start)
	log.Infof("ManuallyProcessAllocateReserveTicketsJob took %s", elapsed)

	return nil
}

func process(db *gorm.DB, ticketRelease models.TicketRelease) error {
	// Start by getting all ticket releases that have not allocated tickets
	log.WithFields(logrus.Fields{
		"id": ticketRelease.ID,
	}).Info("Starting to process ticket release")

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		log.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error beginning transaction: %s", tx.Error.Error())

		return tx.Error
	}

	// This case should never happen since if tickets are allocated then the ticket release is closed
	// but we check anyway
	if !ticketRelease.HasAllocatedTickets {
		log.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Info("Ticket release has not allocated tickets")
		return nil
	}

	// If the ticket release has a promo code then we skip it
	if ticketRelease.PromoCode != nil {
		log.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Info("Ticket release has a promo code")

		return nil
	}

	// availableTickets := ticketRelease.TicketsAvailable
	payWithin := ticketRelease.PayWithin
	// Calculate when ticket should have been paid depending on the when the ticket was updated

	// Get all allocated tickets that isnt a reserve ticket or deleted
	allocatedTickets, err := models.GetAllTicketsToTicketRelease(tx, ticketRelease.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error getting all tickets to ticket release with ID %d: %s", ticketRelease.ID, err.Error())

		return err
	}

	// Get all reserved tickets that isnt deleted
	reservedTickets, err := models.GetAllReserveTicketsToTicketRelease(tx, ticketRelease.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error getting all reserve tickets to ticket release with ID %d: %s", ticketRelease.ID, err.Error())

		return err
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
			log.WithFields(logrus.Fields{
				"id": ticketRelease.ID,
			}).Error("Pay within is nil")

			return errors.New("Pay within is nil")
		}

		if ticket.IsPaid || time.Now().Before(mustPayBefore) {
			// If the ticket has been paid or we are still within the time limit then we continue
			continue
		}

		// If we reach this point then the ticket has not been paid within the time limit
		// We can say that the user looses the ticket, we set the ticket to deleted and
		// increment the number of new tickets to be allocated from the reserve list
		if err := ticket.Delete(tx); err != nil {
			log.WithFields(logrus.Fields{
				"id": ticketRelease.ID,
			}).Errorf("Error deleting ticket with ID %v: %s", ticketRelease.ID, err.Error())

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
					log.WithFields(logrus.Fields{
						"id": ticketRelease.ID,
					}).Errorf("Error saving ticket with ID %d: %s", ticketRelease.ID, err.Error())
				}

				continue
			}

			// We want to allocate the ticket
			// We set the ticket to not be a reserve ticket and set the reserve number to 0
			ticket.IsReserve = false
			ticket.ReserveNumber = 0

			if err := tx.Save(&ticket).Error; err != nil {
				log.WithFields(logrus.Fields{
					"id": ticketRelease.ID,
				}).Errorf("Error saving ticket with ID %d: %s", ticketRelease.ID, err.Error())

				continue
			}

			// TODO Notify user that ticket has been allocated

		}

	}

	err = tx.Commit().Error
	if err != nil {
		log.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error committing transaction: %s", err.Error())

		return err
	}

	log.WithFields(logrus.Fields{
		"id": ticketRelease.ID,
	}).Info("Finished processing ticket release")

	return nil

}
