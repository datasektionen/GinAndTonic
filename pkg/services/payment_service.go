package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/stripe/stripe-go/v72"
	"gorm.io/gorm"
)

type PaymentService struct {
	DB *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{DB: db}
}

func (ps *PaymentService) createPendingTransaction(
	ticketID int,
	eventID int,
	pi *stripe.PaymentIntent,
	user *models.User,
) error {
	// We check if a pending transaction with ticket_id, event_id and user_ug_kth_id already exists
	var existingTransaction models.Transaction
	if err := ps.DB.Where("ticket_id = ? AND event_id = ? AND user_ug_kth_id = ?", ticketID, eventID, user.UGKthID).First(&existingTransaction).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	if existingTransaction.ID != 0 {
		if existingTransaction.Status == models.TransactionStatusCompleted {
			return errors.New("ticket already paid")
		}
		// Delete the existing transaction
		if err := ps.DB.Unscoped().Delete(&existingTransaction).Error; err != nil {
			return err
		}
	}

	// Create the Transaction instance
	transaction := models.Transaction{
		PaymentIntentID: pi.ID,
		TicketID:        ticketID,
		EventID:         eventID,
		Amount:          int(pi.Amount),
		Currency:        pi.Currency,
		Status:          models.TransactionStatusPending,
		UserUGKthID:     user.UGKthID,
		TransactionType: models.TypePurchase,
	}

	if err := ps.DB.Create(&transaction).Error; err != nil {
		return err
	}

	return nil
}

func (ps *PaymentService) ProcessEvent(
	event *stripe.Event,
) *types.ErrorResponse {
	// Implement this function to actually process the event
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: fmt.Sprintf("Error parsing webhook JSON: %v", err.Error())}
		}

		ticketIdstring, ok := paymentIntent.Metadata["tessera_ticket_id"]
		if !ok {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket ID not found in payment intent metadata"}
		}

		ticketId, err := strconv.Atoi(ticketIdstring)
		if err != nil {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Invalid ticket ID"}
		}

		// Start a new transaction
		tx := ps.DB.Begin()

		// If the function returns an error, rollback the transaction
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		ticket, err := HandleSuccessfulTicketPayment(tx, ticketId)
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error handling ticket payment"}
		}

		// Check if existing transaction exists
		var transaction models.Transaction
		if err := tx.Where("ticket_id = ?", ticket.ID).First(&transaction).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				fmt.Println(err)
				return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "An unexpected error occurred, contact event managers"}
			}
		}

		if err := SuccessfulPayment(tx, paymentIntent); err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error handling successful payment"}
		}

		err = Notify_TicketPaymentConfirmation(tx, int(ticket.ID))
		if err != nil {
			fmt.Println(err)
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error notifying user, but ticket payment was successful"}
		}

		// If everything went well, commit the transaction
		txerr := tx.Commit().Error
		if txerr != nil {
			return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error committing transaction"}
		}
	case "payment_intent.created":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: fmt.Sprintf("Error parsing webhook JSON: %v", err.Error())}
		}

		userID := paymentIntent.Metadata["tessera_user_id"]
		var user models.User
		if err := ps.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: fmt.Sprintf("Error finding user: %v", err.Error())}
		}

		ticketIDString, ok := paymentIntent.Metadata["tessera_ticket_id"]
		if !ok {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket ID not found in payment intent metadata"}
		}

		ticketID, err := strconv.Atoi(ticketIDString)
		if err != nil {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Invalid ticket ID"}
		}

		eventIDString, ok := paymentIntent.Metadata["tessera_event_id"]
		if !ok {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Event ID not found in payment intent metadata"}
		}

		eventID, err := strconv.Atoi(eventIDString)
		if err != nil {
			return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Invalid event ID"}
		}

		// Implement the logic to create a pending transaction record.
		// It's a good practice to encapsulate the logic in a method for clarity and reuse.
		err = ps.createPendingTransaction(ticketID, eventID, &paymentIntent, &user)
		if err != nil {
			return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("Error creating pending transaction: %v", err)}
		}
	case "charge.succeeded":
		// Implement the logic to handle a successful charge event
		return nil
	default:
		return &types.ErrorResponse{StatusCode: http.StatusOK, Message: fmt.Sprintf("Unhandled event type: %s", event.Type)}
	}

	return nil
}
