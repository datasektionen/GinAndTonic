package jobs

import (
	"os"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services/allocate_service"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var artd_logger = logrus.New()

func init() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	filePath := "logs/allocate_reserve_tickets_directly_job.log"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.Create(filePath)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		artd_logger.Fatal(err)
	}

	// Set log output to the file
	artd_logger.SetOutput(file)

	// Set log level
	artd_logger.SetLevel(logrus.InfoLevel)

	// Log as JSON for structured logging
	artd_logger.SetFormatter(&logrus.JSONFormatter{})
}

func AllocateReservedTicketsDirectlyJob(db *gorm.DB) error {
	artd_logger.Info("Starting to process reserved ticket requests")
	start := time.Now()

	openAndReservedTicketReleases, err := models.GetOpenReservedTicketReleases(db)
	if err != nil {
		artd_logger.Error(logrus.Fields{
			"error": err,
		}, "Error getting open and reserved ticket releases")
		return err
	}

	if len(openAndReservedTicketReleases) == 0 {
		artd_logger.Info("No open and reserved ticket releases found")
		return nil
	}

	for _, ticketRelease := range openAndReservedTicketReleases {
		if ticketRelease.TicketReleaseMethodDetail.TicketReleaseMethod.MethodName != string(models.RESERVED_TICKET_RELEASE) {
			artd_logger.WithFields(logrus.Fields{
				"ticket_release_id": ticketRelease.ID,
			}).Info("Ticket release is not a reserved ticket release")
			continue
		}

		err := process_artd(db, &ticketRelease)
		if err != nil {
			artd_logger.Error(err)
			// Depending on your error handling strategy, you can decide whether
			// to continue processing the next ticket release or return the error.
			// For now, we're just printing the error and moving to the next ticket release.
		}
	}

	elapsed := time.Since(start)
	artd_logger.Infof("AllocateReserveTicketsJob took %s", elapsed)

	return nil
}

func process_artd(db *gorm.DB, ticketRelease *models.TicketRelease) error {
	// Get all ticket requests that are not handled

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if there is a payment deadline set, otherwise we set it
	_, err := models.CreateReservedTicketReleasePaymentDeadline(tx, ticketRelease.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// This is the total number of tickets available for the ticket release
	var totalTicketsAvailable int = ticketRelease.TicketsAvailable

	allAllocatedTickets, err := models.GetAllTicketsToTicketRelease(tx, ticketRelease.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	var numberOfAlreadyAllocatedTickets int = len(allAllocatedTickets)

	ticketOrders, err := models.GetAllValidTicketOrdersToTicketRelease(tx, ticketRelease.ID)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, ticketOrder := range ticketOrders {
		// Allocate ticket requests directly
		if err != nil {
			tx.Rollback()
			return err
		}

		if totalTicketsAvailable < numberOfAlreadyAllocatedTickets {
			artd_logger.WithFields(logrus.Fields{
				"ticket_release_id": ticketRelease.ID,
			}).Info("No more tickets available for ticket release")
			break
		}

		ticket, err := allocate_service.AllocateTicketOrder(ticketRequest, tx)

		if err != nil {
			tx.Rollback()
			return err
		}

		err = Notify_ReservedTicketAllocated(tx, int(ticket.ID), ticket.PaymentDeadline)

		if err != nil {
			tx.Rollback()
			return err
		}

		artd_logger.WithFields(logrus.Fields{
			"ticket_id":         ticket.ID,
			"ticket_request_id": ticket.TicketRequestID,
		}).Info("Allocated ticket directly")

		numberOfAlreadyAllocatedTickets++
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
