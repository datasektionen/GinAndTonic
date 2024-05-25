package allocate_service

import (
	"errors"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

func AllocateTicket(ticketRequest models.TicketRequest, tx *gorm.DB) (*models.Ticket, error) {
	if ticketRequest.TicketRelease.PaymentDeadline == nil {
		return nil, errors.New("no payment deadline specified")
	}

	paymentDeadline := ticketRequest.TicketRelease.PaymentDeadline

	if ticketRequest.TicketType.ID == 0 {
		// Fatal error, but we can just load the ticket type
		if err := tx.Preload("TicketType").First(&ticketRequest).Error; err != nil {
			return nil, err
		}
	}

	if ticketRequest.TicketType.ID == 0 {
		// Fatal error
		return nil, errors.New("no ticket type specified")
	}

	var isPaid bool = false
	// If the price of the ticket is 0, set it to have been paid
	if ticketRequest.TicketType.Price == 0 && ticketRequest.TicketType.ID != 0 {
		isPaid = true
	}

	var qrCode string = utils.GenerateRandomString(16)
	now := time.Now()

	ticket := models.Ticket{
		TicketRequestID: ticketRequest.ID,
		IsReserve:       false,
		UserUGKthID:     ticketRequest.UserUGKthID,
		IsPaid:          isPaid,
		QrCode:          qrCode,
		PurchasableAt:   &now,
		PaymentDeadline: &paymentDeadline.OriginalDeadline,
	}

	if err := tx.Create(&ticket).Error; err != nil {
		return nil, err
	}

	ticketRequest.IsHandled = true
	if err := tx.Save(&ticketRequest).Error; err != nil {
		return nil, err
	}

	// Set the TicketID in the ticketRequest.TicketAddOn.TicketID to the ticket.ID

	if err := tx.Model(&models.TicketAddOn{}).Where("ticket_request_id = ?", ticketRequest.ID).Update("ticket_id", ticket.ID).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
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
