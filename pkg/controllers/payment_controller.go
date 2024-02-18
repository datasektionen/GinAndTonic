package controllers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/webhook"
	"gorm.io/gorm"
)

// TODO Implement payment log file for better debugging in production

var endpointSecret string

func init() {
	// Set your secret key. Remember to switch to your live secret key in production!
	// See your keys here: https://dashboard.stripe.com/apikeys
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	endpointSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
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

	metadata := map[string]string{
		"tessera_ticket_id":       strconv.Itoa(ticketId),
		"tessera_event_id":        strconv.Itoa(ticket.TicketRequest.TicketRelease.EventID),
		"tessera_event_date":      ticket.TicketRequest.TicketRelease.Event.Date.Format("2006-01-02"),
		"tessera_ticket_type_id":  strconv.Itoa(int(ticket.TicketRequest.TicketTypeID)),
		"tessera_user_id":         user.UGKthID,
		"tessera_recipient_email": user.Email,
		"tessera_event_name":      ticket.TicketRequest.TicketRelease.Event.Name,
		"tessera_ticket_release":  ticket.TicketRequest.TicketRelease.Name,
		"tessera_ticket_type":     ticket.TicketRequest.TicketType.Name,
		"tessera_ticket_amount":   strconv.Itoa(ticket.TicketRequest.TicketAmount),
		"tessera_ticket_price":    fmt.Sprintf("%f", ticket.TicketRequest.TicketType.Price),
	}

	params := &stripe.PaymentIntentParams{
		Params: stripe.Params{
			Metadata: metadata,
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

	idempotencyKey := fmt.Sprintf("payment-intent-%d-%s-%s", ticketId, ugkthid, ticket.TicketRequest.TicketRelease.Event.Name)
	params.IdempotencyKey = stripe.String(idempotencyKey)

	pi, err := paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing webhook: %v", err.Error())
		return
	}

	var webhookEvent models.WebhookEvent
	if err := pc.DB.Where("stripe_id = ?", event.ID).First(&webhookEvent).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusInternalServerError, "Error checking for existing webhook event")
			return
		}
		// If the webhook event does not exist, create a new one
		webhookEvent = models.WebhookEvent{
			StripeID:  event.ID,
			EventType: event.Type,
			Processed: false, // Initially false, will be set to true once processed
		}
		if err := pc.DB.Create(&webhookEvent).Error; err != nil {
			c.String(http.StatusInternalServerError, "Error creating webhook event")
			return
		}
	} else if webhookEvent.Processed {
		// If the webhook event already exists and has been processed, return
		c.String(http.StatusOK, "Webhook event already processed")
		return
	}

	// Process the event
	peErr := pc.pService.ProcessEvent(&event)
	if peErr != nil {
		webhookEvent.LastError = peErr.Message
	} else {
		webhookEvent.Processed = true
	}

	// Save the processed event
	if err := pc.DB.Save(&webhookEvent).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error saving webhook event")
		return
	}

	if peErr != nil {
		c.String(peErr.StatusCode, peErr.Message)
	} else {
		c.Status(http.StatusOK)
	}
}
