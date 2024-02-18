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

		ticketIdstring, ok := paymentIntent.Metadata["ticket_id"]
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
		break
	default:
		return &types.ErrorResponse{StatusCode: http.StatusOK, Message: fmt.Sprintf("Unhandled event type: %s", event.Type)}
	}

	return nil
}
