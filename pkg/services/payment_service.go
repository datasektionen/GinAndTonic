package services

import (
	"encoding/json"
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

func (ps *PaymentService) processEventWithRetries(event *stripe.Event) error {
	const maxRetries = 3
	var lastError error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("Retrying processing event %s, attempt %d\n", event.ID, attempt)
		}

		err := ps.ProcessEvent(event) // Implement this function to actually process the event
		if err != nil {
			lastError = err
			// Update the event in the database with retry count and last error
			ps.DB.Model(&models.WebhookEvent{}).Where("stripe_id = ?", event.ID).Updates(map[string]interface{}{
				"RetryCount": gorm.Expr("retry_count + ?", 1),
				"LastError":  err.Error(),
			})
			continue
		}

		// On successful processing, break the loop
		return nil
	}

	return lastError // Return the last error encountered
}

func (ps *PaymentService) createPendingTransaction(
	ticketID int,
	eventID int,
	pi *stripe.PaymentIntent,
	user *models.User,
) error {
	// Create the Transaction instance
	transaction := models.Transaction{
		PaymentIntentID: &pi.ID,
		TicketID:        &ticketID,
		EventID:         &eventID,
		Amount:          int(pi.Amount),
		Currency:        pi.Currency,
		Status:          models.TransactionStatusPending,
		UserUGKthID:     user.UGKthID,
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

		ticket, err := HandleSuccessfullTicketPayment(tx, ticketId)
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error handling ticket payment"}
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
		if err := ps.DB.Where("ug_kth_id = ?", userID).First(&user).Error; err != nil {
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
