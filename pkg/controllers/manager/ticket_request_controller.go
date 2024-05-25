package manager_controller

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/services"
	manager_service "github.com/DowLucas/gin-ticket-release/pkg/services/manager"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ManagerTicketRequestController struct {
	DB         *gorm.DB
	trService  *services.TicketRequestService
	mtrService *manager_service.ManagerTicketRequestService
}

func NewManagerTicketRequestController(db *gorm.DB) *ManagerTicketRequestController {
	trservice := services.NewTicketRequestService(db)
	mtrequestservice := manager_service.NewManagerTicketRequestService(db)
	return &ManagerTicketRequestController{DB: db, trService: trservice, mtrService: mtrequestservice}
}

/*
This controller handles manager actions made on ticket requests. for instance deleting or un-deleting a ticket request.
*/

type ReqTR struct {
	Action           string `json:"action" binding:"required"`
	TicketRequestIDs []int  `json:"ticket_request_ids"`
}

// DeleteTicketRequest is a method that deletes a ticket request.
func (trc *ManagerTicketRequestController) TicketRequestAction(c *gin.Context) {
	// Body consists of ticekt_request_ids to be deleted

	var req ReqTR
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch req.Action {
	case "delete":
		err := trc.mtrService.DeleteTicketRequests(req.TicketRequestIDs)
		if err != nil {
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			return
		}

		c.JSON(200, gin.H{"message": "Ticket request deleted"})
	case "undelete":
		err := trc.mtrService.UndeleteTicketRequests(req.TicketRequestIDs)
		if err != nil {
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			return
		}
	case "allocate":
		err := services.SelectivelyAllocateTicketRequests(
			trc.DB,
			req.TicketRequestIDs,
		)
		if err != nil {
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
}
