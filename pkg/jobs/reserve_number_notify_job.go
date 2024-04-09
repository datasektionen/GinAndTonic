package jobs

import (
	"os"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Function that notifies reserve tickets users about their new reserve number
var resnum_logger = logrus.New()

func init() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	if _, err := os.Stat("logs/notify_reserve_number_job.log"); os.IsNotExist(err) {
		os.Create("logs/notify_reserve_number_job.log")
	}

	allocator_log_file, err := os.OpenFile("logs/notify_reserve_number_job.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		resnum_logger.Fatal(err)
	}

	// Set log output to the file
	resnum_logger.SetOutput(allocator_log_file)

	// Set log level
	resnum_logger.SetLevel(logrus.InfoLevel)

	// Log as JSON for structured logging
	resnum_logger.SetFormatter(&logrus.JSONFormatter{})
}

func NotifyReserveNumberJob(db *gorm.DB) error {
	start := time.Now()

	closedTicketReleases, err := models.GetClosedTicketReleases(db)
	if err != nil {
		return err
	}

	if len(closedTicketReleases) == 0 {
		resnum_logger.Info("No closed ticket releases")
		return nil
	}

	resnum_logger.WithFields(logrus.Fields{
		"number_of_closed_ticket_releases": len(closedTicketReleases),
	}).Info("Starting to process closed ticket releases")

	for _, ticketRelease := range closedTicketReleases {
		err := process_ntnj(db, &ticketRelease)
		if err != nil {
			// Depending on your error handling strategy, you can decide whether
			// to continue processing the next ticket release or return the error.
			// For now, we're just printing the error and moving to the next ticket release.
		}
	}

	elapsed := time.Since(start)
	resnum_logger.Infof("AllocateReserveTicketsJob took %s", elapsed)

	return nil
}

func process_ntnj(db *gorm.DB, ticketRelease *models.TicketRelease) error {
	tx := db.Begin()
	if tx.Error != nil {
		resnum_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error beginning transaction: %s", tx.Error.Error())
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil || tx.Error != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		resnum_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error beginning transaction: %s", tx.Error.Error())

		return tx.Error
	}

	// Check if the ticket release has been allocated
	if !ticketRelease.HasAllocatedTickets {
		tx.Rollback()
		return nil
	}

	// Get all reserved tickets that aren't deleted
	reservedTickets, err := models.GetAllReserveTicketsToTicketRelease(tx, ticketRelease.ID)
	if err != nil {
		resnum_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Errorf("Error getting all reserve tickets to ticket release with ID %d: %s", ticketRelease.ID, err.Error())
		tx.Rollback()
		return err
	}

	// Check if there are any reserved tickets
	if len(reservedTickets) == 0 {
		resnum_logger.WithFields(logrus.Fields{
			"id": ticketRelease.ID,
		}).Info("No reserve tickets")
	}

	resnum_logger.WithFields(logrus.Fields{
		"number_of_reserved_tickets": len(reservedTickets),
	}).Info("Starting to process reserved tickets")

	for _, ticket := range reservedTickets {
		// Get the user that reserved the ticket
		if err != nil {
			resnum_logger.WithFields(logrus.Fields{
				"id": ticket.ID,
			}).Errorf("Error getting user with ID %s: %s", ticket.User.UGKthID, err.Error())

			return err
		}

		if ticket.ReserveNumber == 0 {
			// The user is going to get a ticket so no need to notify them
			continue
		}

		// Notify the user about their reserve number
		err = Notify_UpdateReserveNumbers(tx, int(ticket.ID))
		if err != nil {
			resnum_logger.WithFields(logrus.Fields{
				"id": ticket.ID,
			}).Errorf("Error notifying user with ID %s about their reserve number: %s", ticket.User.UGKthID, err.Error())

			continue
		}
	}

	return tx.Commit().Error
}
