package manager_controller

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/services"
	manager_service "github.com/DowLucas/gin-ticket-release/pkg/services/manager"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ManagerTicketOrderController struct {
	DB         *gorm.DB
	service    *services.TicketOrderService
	mtrService *manager_service.ManagerTicketOrderService
}

func NewManagerticketOrderController(db *gorm.DB) *ManagerTicketOrderController {
	service := services.NewTicketOrderService(db)
	mtrequestservice := manager_service.NewManagerTicketOrderService(db)
	return &ManagerTicketOrderController{DB: db, service: service, mtrService: mtrequestservice}
}

/*
This controller handles manager actions made on ticket requests. for instance deleting or un-deleting a ticket request.
*/

type ReqTR struct {
	Action         string `json:"action" binding:"required"`
	TicketOrderIDs []int  `json:"ticket_order_ids"`
}

// DeleteticketOrder is a method that deletes a ticket request.
func (trc *ManagerTicketOrderController) TicketOrderAction(c *gin.Context) {
	// Body consists of ticekt_request_ids to be deleted

	var req ReqTR
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch req.Action {
	case "delete":
		err := trc.mtrService.DeleteTicketOrder(req.TicketOrderIDs)
		if err != nil {
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			return
		}

		c.JSON(200, gin.H{"message": "Ticket request deleted"})
	case "undelete":
		err := trc.mtrService.UndeleteTicketOrders(req.TicketOrderIDs)
		if err != nil {
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			return
		}
	case "allocate":
		err := services.SelectivelyAllocateTicketOrders(
			trc.DB,
			req.TicketOrderIDs,
		)
		if err != nil {
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
}
