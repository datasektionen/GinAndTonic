package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketRequestController struct {
	Service *services.TicketRequestService
}

func NewTicketRequestController(db *gorm.DB) *TicketRequestController {
	service := services.NewTicketRequestService(db)
	return &TicketRequestController{Service: service}
}

func (trc *TicketRequestController) UsersList(c *gin.Context) {
	// Find all ticket requests for the user

	UGKthId, _ := c.Get("ugkthid")

	ticketRequests, err := trc.Service.GetTicketRequests(UGKthId.(string))

	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ticket_requests": ticketRequests})
}

func (trc *TicketRequestController) Create(c *gin.Context) {
	var ticketRequests []models.TicketRequest

	UGKthID, _ := c.Get("ugkthid")

	if err := c.ShouldBindJSON(&ticketRequests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i := range ticketRequests {
		ticketRequests[i].UserUGKthID = UGKthID.(string)
	}

	err := trc.Service.CreateTicketRequests(ticketRequests)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusCreated, ticketRequests)
}
func (trc *TicketRequestController) Get(c *gin.Context) {
	UGKthID, _ := c.Get("ugkthid")
	ticketRequests, err := trc.Service.GetTicketRequests(UGKthID.(string))

	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusOK, ticketRequests)
}
