package allocate_service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

func AllocateTicket(ticket models.Ticket, paymentDeadline models.TicketReleasePaymentDeadline, tx *gorm.DB) (*models.Ticket, error) {
	if ticket.IsHandled {
		return &ticket, nil
	}

	if ticket.TicketType.ID == 0 {
		// Fatal error
		return nil, errors.New("no ticket type specified")
	}

	var isPaid bool = false
	// If the price of the ticket is 0, set it to have been paid
	if ticket.TicketType.Price == 0 && ticket.TicketType.ID != 0 {
		isPaid = true
	}

	var qrCode string = utils.GenerateRandomString(16)
	now := time.Now()

	ticket.IsHandled = true
	ticket.IsPaid = isPaid
	ticket.QrCode = qrCode
	ticket.PurchasableAt = sql.NullTime{Time: now, Valid: true}
	ticket.PaymentDeadline = sql.NullTime{Time: paymentDeadline.OriginalDeadline, Valid: true}

	if err := tx.Save(&ticket).Error; err != nil {
		return nil, err
	}

	// Set the TicketID in the ticketRequest.TicketAddOn.TicketID to the ticket.ID

	return &ticket, nil
}

func AllocateTicketOrder(ticketOrder models.TicketOrder, tx *gorm.DB) (*[]models.Ticket, error) {
	if ticketOrder.IsTicketRequest() {
		return &ticketOrder.Tickets, nil
	}

	paymentDeadline := ticketOrder.TicketRelease.PaymentDeadline

	var tickets []models.Ticket

	for _, ticket := range ticketOrder.Tickets {
		ticket, err := AllocateTicket(ticket, paymentDeadline, tx)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, *ticket)
	}

	ticketOrder.Type = models.TicketOrderTicket
	if err := tx.Save(&ticketOrder).Error; err != nil {
		return nil, err
	}

	return &tickets, nil
}

func AllocateReserveTicket(
	ticketRequest models.TicketRequest,
	reserveNumber uint,
	tx *gorm.DB) (*models.Ticket, error) {
	ticketRequest.IsHandled = true
	if err := tx.Save(&ticketRequest).Error; err != nil {
		return nil, err
	}

	qrCode := utils.GenerateRandomString(16)

	ticket := models.Ticket{
		TicketRequestID: ticketRequest.ID,
		ReserveNumber:   reserveNumber,
		IsReserve:       true,
		UserUGKthID:     ticketRequest.UserUGKthID,
		QrCode:          qrCode,
	}

	if err := tx.Create(&ticket).Error; err != nil {
		return nil, err
	}

	// S	et the TicketID in the ticketRequest.TicketAddOn.TicketID to the ticket.ID
	if err := tx.Model(&models.TicketAddOn{}).Where("ticket_request_id = ?", ticketRequest.ID).Update("ticket_id", ticket.ID).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}
