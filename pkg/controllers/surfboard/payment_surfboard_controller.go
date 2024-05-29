package surfboard_controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	surfboard_service_order "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/order"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentSurfboardController struct {
	DB                 *gorm.DB
	createOrderService *surfboard_service_order.SurfboardCreateOrderService
	orderService       *surfboard_service_order.OrderService
}

func NewPaymentSurfboardController(db *gorm.DB) *PaymentSurfboardController {
	return &PaymentSurfboardController{DB: db,
		createOrderService: surfboard_service_order.NewSurfboardCreateOrderService(db),
		orderService:       surfboard_service_order.NewOrderService()}
}

type CreateOrderRequest struct {
	TicketIDs []uint `json:"ticket_ids"`
}

func (psc *PaymentSurfboardController) CreateOrder(c *gin.Context) {
	// This is the initial entry point when a user want to pay for a ticket or multiple tickets.
	// The user will send a list of ticket IDs that they want to pay for.

	user := c.MustGet("user").(models.User)
	eventRef := c.Param("referenceID")

	var event models.Event
	if err := psc.DB.Preload("Terminal").Where("reference_id = ?", eventRef).First(&event).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event not found"})
		return
	}

	network, err := event.GetNetwork(psc.DB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Network not found"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the order
	order, err := psc.createOrderService.CreateOrder(req.TicketIDs, network, &event, &user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return a redirect to the payment page
	c.JSON(http.StatusOK, gin.H{"order": order, "message": "Order created"})
}

func (psc *PaymentSurfboardController) GetOrderStatus(c *gin.Context) {
	// This endpoint is used to check the status of an order
	// The user will send an order ID and the endpoint will return the status of the order
	orderID := c.Param("orderID")

	var order models.Order
	if err := psc.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order not found"})
		return
	}

	status, err := psc.orderService.GetOrderStatus(&order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order, "status": status})
}
