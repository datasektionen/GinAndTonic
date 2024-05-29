package surfboard_controllers

import (
	surfboard_service_order "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/order"
	"gorm.io/gorm"
)

type PaymentSurfboardController struct {
	DB           *gorm.DB
	orderService *surfboard_service_order.SurfboardOrderService
}

func NewPaymentSurfboardController(db *gorm.DB) *PaymentSurfboardController {
	return &PaymentSurfboardController{DB: db, orderService: surfboard_service_order.NewSurfboardOrderService(db)}
}

type CreateOrderRequest struct {
	TicketIDs []uint `json:"ticket_ids"`
}

// func (psc *PaymentSurfboardController) CreateOrder(c *gin.Context) {
// 	// This is the initial entry point when a user want to pay for a ticket or multiple tickets.
// 	// The user will send a list of ticket IDs that they want to pay for.

// 	user := c.MustGet("user").(models.User)

// 	var req CreateOrderRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Create the order
// 	order, err := psc.orderService.CreateOrder(req.TicketIDs, &user)

// }
