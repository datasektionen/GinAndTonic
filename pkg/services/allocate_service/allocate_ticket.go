package allocate_service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

func AllocateTicket(ticket *models.Ticket, paymentDeadline models.TicketReleasePaymentDeadline, tx *gorm.DB) error {
	if ticket.TicketOrder.IsHandled {
		return errors.New("ticket order is already handled")
	}

	if ticket.TicketType.ID == 0 {
		// Fatal error
		return errors.New("no ticket type specified")
	}

	var isPaid bool = false
	// If the price of the ticket is 0, set it to have been paid
	if ticket.TicketType.Price == 0 && ticket.TicketType.ID != 0 {
		isPaid = true
	}

	var qrCode string = utils.GenerateRandomString(16)
	now := time.Now()

	ticket.IsPaid = isPaid
	ticket.QrCode = qrCode
	ticket.IsReserve = false
	ticket.PurchasableAt = sql.NullTime{Time: now, Valid: true}
	ticket.PaymentDeadline = sql.NullTime{Time: paymentDeadline.OriginalDeadline, Valid: true}
	ticket.Status = models.Allocated

	if err := tx.Save(&ticket).Error; err != nil {
		return err
	}

	// Set ticket.TicketOrder.IsHandled to true
	if err := tx.Model(&ticket.TicketOrder).Update("is_handled", true).Error; err != nil {
		return err
	}

	// Set the TicketID in the ticketOrder.TicketAddOn.TicketID to the ticket.ID

	return nil
}

func AllocateTicketOrder(ticketOrder models.TicketOrder, tx *gorm.DB) (*[]models.Ticket, error) {
	if ticketOrder.IsticketOrder() {
		return &ticketOrder.Tickets, nil
	}

	paymentDeadline := ticketOrder.TicketRelease.PaymentDeadline

	var tickets []models.Ticket

	for _, ticket := range ticketOrder.Tickets {
		err := AllocateTicket(&ticket, paymentDeadline, tx)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	ticketOrder.Type = models.TicketOrderTicket
	if err := tx.Save(&ticketOrder).Error; err != nil {
		return nil, err
	}

	return &tickets, nil
}

func AllocateReserveTicket(
	ticket *models.Ticket,
	reserveNumber uint,
	tx *gorm.DB) error {
	if ticket.TicketOrder.IsHandled {
		return nil
	}

	qrCode := utils.GenerateRandomString(16)

	ticket.IsPaid = false
	ticket.QrCode = qrCode
	ticket.IsReserve = true
	ticket.ReserveNumber = reserveNumber
	ticket.Status = models.Reserve

	if err := tx.Save(&ticket).Error; err != nil {
		return err
	}

	if err := tx.Model(&ticket.TicketOrder).Update("is_handled", true).Error; err != nil {
		return err
	}

	return nil
}
