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

func (trc *TicketRequestController) Create(c *gin.Context) {
	var ticketRequest models.TicketRequest

	UGKthID, _ := c.Get("ugkthid")
	ticketRequest.UserUGKthID = UGKthID.(string)

	if err := c.ShouldBindJSON(&ticketRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := trc.Service.CreateTicketRequest(&ticketRequest)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusCreated, ticketRequest)
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
