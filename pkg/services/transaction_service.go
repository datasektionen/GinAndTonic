package services

import (
	"fmt"
	"strconv"
	"time"

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

func (ts *TransactionService) CreateTransaction(
	pi stripe.PaymentIntent,
	ticket *models.Ticket,
	eventId int,
	transactionStatus models.TransactionStatus,
) error {
	// Extracting the ticket ID from metadata
	ticketIDStr, ok := pi.Metadata["tessera_ticket_id"]
	if !ok {
		return fmt.Errorf("ticket_id not found in payment intent metadata")
	}
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		return err
	}

	// Create the Transaction instance
	transaction := models.Transaction{
		PaymentIntentID: pi.ID,
		EventID:         eventId,
		TicketID:        ticketID,
		Amount:          int(pi.Amount),
		Currency:        pi.Currency,
		Refunded:        false, // Set this based on your logic
		UserUGKthID:     ticket.UserUGKthID,
		Status:          transactionStatus,
	}

	if err := transaction.Validate(); err != nil {
		return err
	}

	// Assuming you have a way to save the Transaction in your service
	if err := ts.DB.Create(&transaction).Error; err != nil {
		return err
	}

	return nil
}

func SuccessfulPayment(
	db *gorm.DB, // Allows transaction to be passed in
	pi stripe.PaymentIntent) error {
	var transaction models.Transaction
	if err := db.Where("payment_intent_id = ?", pi.ID).Find(&transaction).Error; err != nil {
		return err
	}

	now := time.Now().Unix()

	transaction.Status = models.TransactionStatusCompleted
	transaction.PayedAt = &now
	// transaction.PaymentMethod = &pm.Type // TODO Implement
	transaction.TransactionType = models.TypePurchase

	if err := db.Save(&transaction).Error; err != nil {
		return err
	}

	return nil
}
