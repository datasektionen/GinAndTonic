package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/webhook"
	"gorm.io/gorm"
)

// TODO Implement payment log file for better debugging in production

func init() {
	// Set your secret key. Remember to switch to your live secret key in production!
	// See your keys here: https://dashboard.stripe.com/apikeys
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

type PaymentController struct {
	DB                 *gorm.DB
	tpService          *services.TicketPaymentService
	pService           *services.PaymentService
	transactionService *services.TransactionService
}

func NewPaymentController(db *gorm.DB) *PaymentController {
	pService := services.NewPaymentService(db)
	return &PaymentController{DB: db,
		tpService:          services.NewTicketPaymentService(db),
		transactionService: services.NewTransactionService(db),
		pService:           pService}
}

func (pc *PaymentController) CreatePaymentIntent(c *gin.Context) {
	ugkthid := c.MustGet("ugkthid").(string)

	ticketIdString := c.Param("ticketID")
	ticketId, err := strconv.Atoi(ticketIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ticket models.Ticket

	if err := pc.DB.
		Preload("TicketRequest.TicketType").
		Preload("TicketRequest.User").
		Preload("TicketRequest.TicketRelease.Event").
		Where("id = ? AND user_ug_kth_id = ?", ticketId, ugkthid).First(&ticket).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := ticket.TicketRequest.User

	var paymentIntentID string
	var existingTransaction models.Transaction

	// Attempt to find an existing, unused payment intent for the ticket and user
	if err := pc.DB.Where("ticket_id = ? AND user_ug_kth_id = ?", ticketId, ugkthid).First(&existingTransaction).Error; err == nil {
		paymentIntentID = *existingTransaction.PaymentIntentID
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve existing transaction"})
		return
	}

	pi, er := handleExistingTransactionAndPaymentIntent(pc.DB, paymentIntentID, &existingTransaction)

	if er != nil {
		c.JSON(er.StatusCode, gin.H{"error": er.Message})
		return
	} else {
		if pi != nil {
			c.JSON(http.StatusOK, gin.H{"client_secret": pi.ClientSecret})
			return
		}
	}

	// Sum price
	var totalPrice int64
	totalPrice += (int64)(ticket.TicketRequest.TicketType.Price*100) * (int64)(ticket.TicketRequest.TicketAmount)

	// Define the customer parameters
	customerParams := &stripe.CustomerListParams{}
	customerParams.Filters.AddFilter("email", "", user.Email)
	customerParams.Single = true

	// Try to find existing customer
	existingCustomerIter := customer.List(customerParams)
	var cust *stripe.Customer
	for existingCustomerIter.Next() {
		cust = existingCustomerIter.Customer()
	}

	if cust == nil {
		// No customer found, creating a new one
		newCustomerParams := &stripe.CustomerParams{
			Email: stripe.String(user.Email),
		}
		newCust, err := customer.New(newCustomerParams)
		if err != nil {
			fmt.Println("Customer creation failed:", err)
			return
		}
		cust = newCust
	}
	params := &stripe.PaymentIntentParams{
		Params: stripe.Params{
			Metadata: map[string]string{
				"tessera_ticket_id":       strconv.Itoa(ticketId),
				"tessera_event_id":        strconv.Itoa(ticket.TicketRequest.TicketRelease.EventID),
				"tessera_event_date":      ticket.TicketRequest.TicketRelease.Event.Date.Format("2006-01-02"), // Assuming Event.Date is a time.Time
				"tessera_ticket_type_id":  strconv.Itoa(int(ticket.TicketRequest.TicketTypeID)),
				"tessera_user_id":         user.UGKthID,
				"tessera_recipient_email": ticket.TicketRequest.User.Email,
				"tessera_event_name":      ticket.TicketRequest.TicketRelease.Event.Name,
				"tessera_ticket_release":  ticket.TicketRequest.TicketRelease.Name,
				"tessera_ticket_type":     ticket.TicketRequest.TicketType.Name,
				"tessera_ticket_amount":   strconv.Itoa(ticket.TicketRequest.TicketAmount),
				"tessera_ticket_price":    fmt.Sprintf("%f", ticket.TicketRequest.TicketType.Price),
			},
		},
		Customer:           stripe.String(cust.ID),
		Amount:             stripe.Int64(totalPrice),
		Currency:           stripe.String(string(stripe.CurrencySEK)),
		PaymentMethodTypes: []*string{stripe.String("card")},
		ReceiptEmail:       stripe.String(ticket.TicketRequest.User.Email),
		Description: stripe.String(fmt.Sprintf("Event Name: %s, Ticket Type: %s",
			ticket.TicketRequest.TicketRelease.Event.Name,
			ticket.TicketRequest.TicketType.Name)),
	}

	idempotencyKey := fmt.Sprintf("payment-intent-%d-%s", ticketId, ugkthid) // Just an example, ensure it's unique per intent
	params.IdempotencyKey = stripe.String(idempotencyKey)
	params.AddMetadata("ticket_id", strconv.Itoa(ticketId))

	pi, err = paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if pc.transactionService.CreateTransaction(*pi,
		&ticket,
		ticket.TicketRequest.TicketRelease.EventID,
		models.TransactionStatusPending) != nil {
		c.String(http.StatusInternalServerError, "Error creating transaction")
		return
	}

	c.JSON(http.StatusOK, gin.H{"client_secret": pi.ClientSecret})
}

// Payment webhook
func (pc *PaymentController) PaymentWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusServiceUnavailable, "Error reading request body: %v", err)
		return
	}

	// This is your Stripe CLI webhook secret for testing your endpoint locally.
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Pass the request body and Stripe-Signature header to ConstructEvent, along
	// with the webhook signing key.
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)

	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing webhook: %v", err.Error())
		return
	}

	var existingWebhookEvent models.WebhookEvent
	if err := pc.DB.Where("stripe_event_id = ? AND processed = ?", event.ID, true).First(&existingWebhookEvent).Error; err == nil {
		c.String(http.StatusOK, "Webhook event already processed")
		return
	} else if err != gorm.ErrRecordNotFound {
		c.String(http.StatusInternalServerError, "Error retrieving existing webhook event")
		return
	}

	webhookEvent := models.WebhookEvent{
		StripeID:  event.ID,
		EventType: event.Type,
		Processed: false, // Initially false, will be set to true once processed
	}
	pc.DB.Create(&webhookEvent)

	// Unmarshal the event data into an appropriate struct depending on its Type
	peErr := pc.pService.ProcessEvent(&event)
	if peErr != nil {
		webhookEvent.LastError = peErr.Message

		if err := pc.DB.Save(&webhookEvent).Error; err != nil {
			c.String(http.StatusInternalServerError, "Error saving webhook event")
			return
		}

		c.String(peErr.StatusCode, peErr.Message)
		return
	}

	c.Status(http.StatusOK)
}

func handleExistingTransactionAndPaymentIntent(
	DB *gorm.DB,
	paymentIntentID string,
	transaction *models.Transaction,
) (pi *stripe.PaymentIntent, err *types.ErrorResponse) {
	if len(paymentIntentID) > 0 {
		// Use the Stripe Go SDK to retrieve the existing payment intent
		pi, err := paymentintent.Get(paymentIntentID, nil)
		if err != nil {
			// Cast the error to a stripe.Error to access more details
			if stripeErr, ok := err.(*stripe.Error); ok {
				switch stripeErr.Type {
				case stripe.ErrorTypeAPIConnection:
					// Handle network/connection errors
					return nil, &types.ErrorResponse{StatusCode: http.StatusServiceUnavailable, Message: "Network error connecting to Stripe"}
				case stripe.ErrorTypeInvalidRequest:
					// Handle invalid requests, such as requesting a non-existent payment intent
					if strings.Contains(stripeErr.Msg, "No such payment_intent") {
						// If the payment intent does not exist, you may choose to create a new one
						// Insert code to create a new PaymentIntent here
						// Delete the existing transaction
						if err := DB.Delete(&transaction).Error; err != nil {
							return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Failed to delete existing transaction"}
						}

						return nil, nil
					} else {
						// Other types of invalid requests
						return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: stripeErr.Msg}
					}
				case stripe.ErrorTypeAPI:
					// Handle API errors that occur on Stripe's side
					return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Stripe API error"}
				case stripe.ErrorTypeAuthentication:
					// Handle authentication errors
					return nil, &types.ErrorResponse{StatusCode: http.StatusUnauthorized, Message: "Authentication with Stripe failed"}
				case stripe.ErrorTypeRateLimit:
					// Handle rate limit errors
					return nil, &types.ErrorResponse{StatusCode: http.StatusTooManyRequests, Message: "Rate limit exceeded with Stripe"}
				default:
					// Handle other errors
					return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "An unknown error occurred with Stripe"}
				}
			} else {
				// Handle any non-Stripe errors
				return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "An error occurred retrieving the payment intent"}
			}
		}

		return pi, nil
	}

	return nil, nil
}
