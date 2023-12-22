package services

import (
	"fmt"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/stripe/stripe-go/v72"
	"gorm.io/gorm"
)

type TransactionService struct {
	DB *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{DB: db}
}

func (ts *TransactionService) CreateTransaction(pi stripe.PaymentIntent, ticket *models.Ticket) error {
	// Extracting the ticket ID from metadata
	ticketIDStr, ok := pi.Metadata["ticket_id"]
	if !ok {
		return fmt.Errorf("ticket_id not found in payment intent metadata")
	}
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		return err
	}

	// Create the Transaction instance
	transaction := models.Transaction{
		TicketID:    uint(ticketID),
		Amount:      int(pi.Amount),
		Currency:    pi.Currency,
		PayedAt:     pi.Created,
		Refunded:    false, // Set this based on your logic
		UserUGKthID: ticket.UserUGKthID,
	}

	// Assuming you have a way to save the Transaction in your service
	if err := ts.DB.Create(&transaction).Error; err != nil {
		return err
	}

	return nil
}
