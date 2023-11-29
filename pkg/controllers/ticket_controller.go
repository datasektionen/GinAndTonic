package controllers

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketController struct {
	DB      *gorm.DB
	Service *services.TicketService
}

func NewTicketController(db *gorm.DB) *TicketController {
	service := services.NewTicketService(db)
	return &TicketController{DB: db, Service: service}
}

func (tc *TicketController) ListTickets(c *gin.Context) {

	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tickets, err := tc.Service.GetAllTickets(eventID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	println("Tickets: ", tickets)

	c.JSON(http.StatusOK, tickets)
}
