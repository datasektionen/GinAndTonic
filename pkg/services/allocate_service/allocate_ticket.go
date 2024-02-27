package allocate_service

import (
	"errors"
	"fmt"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

func AllocateTicket(ticketRequest models.TicketRequest, tx *gorm.DB) (*models.Ticket, error) {
	ticketRequest.IsHandled = true
	if err := tx.Save(&ticketRequest).Error; err != nil {
		return nil, err
	}

	// If the price of the ticket is 0, set it to have been paid
	if ticketRequest.TicketType.ID == 0 {
		// Fatal error
		fmt.Println("No ticket type specified")
		return nil, errors.New("no ticket type specified")
	}

	var isPaid bool = false
	if ticketRequest.TicketType.Price == 0 {
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
	}

	if err := tx.Create(&ticket).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}
