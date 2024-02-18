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

func HandleSuccessfullTicketPayment(
	db *gorm.DB, // Allows transaction to be passed in
	ticketId int,
) (ticket *models.Ticket, err error) {
	// Handles a successfull ticket payment
	if err := db.Preload("TicketRequest").Where("id = ?", ticketId).First(&ticket).Error; err != nil {
		return nil, err
	}

	ticket.IsPaid = true

	if err := db.Save(&ticket).Error; err != nil {
		return nil, err
	}

	return ticket, nil
}
