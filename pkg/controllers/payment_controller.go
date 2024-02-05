package controllers

import (
	"encoding/json"
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
	transactionService *services.TransactionService
}

func NewPaymentController(db *gorm.DB) *PaymentController {
	return &PaymentController{DB: db, tpService: services.NewTicketPaymentService(db), transactionService: services.NewTransactionService(db)}
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

	params := &stripe.PaymentIntentParams{
		Params: stripe.Params{
			Metadata: map[string]string{
				"ticket_id":       strconv.Itoa(ticketId),
				"recipient_email": ticket.TicketRequest.User.Email,
				"event_name":      ticket.TicketRequest.TicketRelease.Event.Name,
				"ticket_release":  ticket.TicketRequest.TicketRelease.Name,
				"ticket_type":     ticket.TicketRequest.TicketType.Name,
				"ticket_amount":   strconv.Itoa(ticket.TicketRequest.TicketAmount),
				"ticket_price":    fmt.Sprintf("%f", ticket.TicketRequest.TicketType.Price),
			},
		},
		Customer:           stripe.String(cust.ID),
		Amount:             stripe.Int64(totalPrice),
		Currency:           stripe.String(string(stripe.CurrencySEK)),
		PaymentMethodTypes: []*string{stripe.String("card")},
		ReceiptEmail:       stripe.String(ticket.TicketRequest.User.Email),
		Description: stripe.String(fmt.Sprintf("Event Name: %s, Ticket Type: %s",
			ticket.TicketRequest.TicketType.Name,
			ticket.TicketRequest.TicketType.Name)),
	}

	params.AddMetadata("ticket_id", strconv.Itoa(ticketId))

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

	// This is your Stripe CLI webhook secret for testing your endpoint locally.
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Pass the request body and Stripe-Signature header to ConstructEvent, along
	// with the webhook signing key.
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)

	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing webhook: %v", err.Error())
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			c.String(http.StatusBadRequest, "Error parsing webhook JSON: %v", err.Error())
			return
		}

		ticketIdstring, ok := paymentIntent.Metadata["ticket_id"]
		if !ok {
			c.String(http.StatusBadRequest, "Ticket ID not found in payment intent metadata")
			return
		}

		ticketId, err := strconv.Atoi(ticketIdstring)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid ticket ID")
			return
		}

		ticket, err := pc.tpService.HandleSuccessfullTicketPayment(ticketId)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error handling ticket payment")
			return
		}

		if pc.transactionService.CreateTransaction(paymentIntent, ticket) != nil {
			c.String(http.StatusInternalServerError, "Error creating transaction")
			return
		}

		err = services.Notify_TicketPaymentConfirmation(pc.DB, int(ticket.ID))
		if err != nil {
			fmt.Println(err)
			c.String(http.StatusInternalServerError, "Error notifying user, but ticket payment was successful")
			return
		}

		// Send notification to user

		// Then define and call a function to handle the event payment_intent.succeeded
		// ...
	// ... handle other event types
	default:
		c.String(http.StatusOK, "Unhandled event type: %s", event.Type)
	}

	c.Status(http.StatusOK)
}
