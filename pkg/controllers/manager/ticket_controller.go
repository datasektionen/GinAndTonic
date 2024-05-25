package manager_controller

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/services"
	manager_service "github.com/DowLucas/gin-ticket-release/pkg/services/manager"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ManagerTicketController struct {
	DB         *gorm.DB
	trService  *services.TicketService
	mtrService *manager_service.ManagerTicketService
}

func NewManagerTicketController(db *gorm.DB) *ManagerTicketController {
	trservice := services.NewTicketService(db)
	mtrequestservice := manager_service.NewManagerTicketService(db)
	return &ManagerTicketController{DB: db, trService: trservice, mtrService: mtrequestservice}
}

/*
This controller handles manager actions made on ticket requests. for instance deleting or un-deleting a ticket request.
*/

type ReqT struct {
	Action    string `json:"action" binding:"required"`
	TicketIDs []int  `json:"ticket_ids"`
}

// DeleteTicket is a method that deletes a ticket request.
func (trc *ManagerTicketController) TicketAction(c *gin.Context) {
	var req ReqT
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch req.Action {
	case "delete":
		err := trc.mtrService.DeleteTickets(req.TicketIDs)
		if err != nil {
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			return
		}

		c.JSON(200, gin.H{"message": "Ticket request deleted"})
	case "undelete":
		err := trc.mtrService.UndeleteTickets(req.TicketIDs)
		if err != nil {
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
}
