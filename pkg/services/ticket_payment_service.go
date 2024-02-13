package services

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type TicketPaymentService struct {
	DB *gorm.DB
}

func NewTicketPaymentService(db *gorm.DB) *TicketPaymentService {
	return &TicketPaymentService{DB: db}
}

func (tps *TicketPaymentService) HandleSuccessfullTicketPayment(
	ticketId int,
) (ticket *models.Ticket, err error) {
	// Handles a successfull ticket payment
	if err := tps.DB.Preload("TicketRequest").Where("id = ?", ticketId).First(&ticket).Error; err != nil {
		return nil, err
	}

	ticket.IsPaid = true

	if err := tps.DB.Save(&ticket).Error; err != nil {
		return nil, err
	}

	return ticket, nil
}
