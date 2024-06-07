package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketOrderController struct {
	Service *services.TicketOrderService
}

func NewTicketOrderController(db *gorm.DB) *TicketOrderController {
	service := services.NewTicketOrderService(db)
	return &TicketOrderController{Service: service}
}

func (trc *TicketOrderController) UsersList(c *gin.Context) {
	UGKthId, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "User ID not found",
		})
		return
	}

	idsString := c.Query("ids")
	var idsInt []int
	if idsString != "" {
		ids := strings.Split(idsString, ",")
		for _, id := range ids {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid ID",
				})
				return
			}
			idsInt = append(idsInt, idInt)
		}
	}

	ticketOrders, err := trc.Service.GetTicketOrdersForUser(UGKthId.(string), &idsInt)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, gin.H{
			"status":  "error",
			"message": fmt.Sprintf(err.Message),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"ticket_orders": ticketOrders,
		},
		"message": "Ticket orders fetched successfully",
	})
}

type TicketOrderCreateRequest struct {
	TicketOrder    models.TicketOrder     `json:"ticket_order"`
	SelectedAddOns []types.SelectedAddOns `json:"selected_add_ons"`
}

// Create a ticket request
func (trc *TicketOrderController) Create(c *gin.Context) {
	var request TicketOrderCreateRequest

	UGKthID, _ := c.Get("user_id")

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketOrder := request.TicketOrder
	ticketOrder.UserUGKthID = UGKthID.(string)

	mTicketOrder, err := trc.Service.CreateTicketOrder(ticketOrder, &request.SelectedAddOns)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	services.Notify_TicketOrderCreated(trc.Service.DB, int(ticketOrder.ID))

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    gin.H{"ticket_order": mTicketOrder},
		"message": "Ticket order created successfully",
	})
}

type TicketOrderGuestCreateRequest struct {
	TicketOrder    models.TicketOrder     `json:"ticket_order"`
	SelectedAddOns []types.SelectedAddOns `json:"selected_add_ons"`
	RequestToken   string                 `json:"request_token" binding:"required"`
}

// Guest reqeust controller
func (trc *TicketOrderController) GuestCreate(c *gin.Context) {
	var request TicketOrderGuestCreateRequest

	userUGKTHId := c.Param("ugkthid")
	if userUGKTHId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the ueer base on the RequestToken
	var user models.User
	if err := trc.Service.DB.Where("id = ? AND request_token = ?", userUGKTHId, request.RequestToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid request token"})
		return
	}

	ticketOrder := request.TicketOrder
	ticketOrder.UserUGKthID = userUGKTHId
	ticketReleaseID := ticketOrder.TicketReleaseID

	var ticketRelease models.TicketRelease

	if err := trc.Service.DB.Preload("TicketReleaseMethodDetail.TicketReleaseMethod").First(&ticketRelease, ticketReleaseID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID"})
		return
	}

	if ticketRelease.TicketReleaseMethodDetail.TicketReleaseMethod.RequiresCustomerAccount() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket release requires customer account"})
		return
	}

	mTicketOrder, err := trc.Service.CreateTicketOrder(ticketOrder, &request.SelectedAddOns)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	services.Notify_GuestTicketOrderCreated(trc.Service.DB, int(ticketOrder.ID))

	c.JSON(http.StatusCreated, gin.H{
		"ticket_order": mTicketOrder,
	})
}

func (trc *TicketOrderController) Get(c *gin.Context) {
	UGKthID, _ := c.Get("user_id")
	ticketOrders, err := trc.Service.GetTicketOrdersForUser(UGKthID.(string), nil)

	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ticket_orders": ticketOrders,
	})
}

func (trc *TicketOrderController) CancelTicketOrder(c *gin.Context) {
	// Get the ID of the ticket request from the URL parameters
	ticketOrderID := c.Param("ticketOrderID")

	// Use your database or service layer to find the ticket request by ID and cancel it
	err := trc.Service.CancelTicketOrder(ticketOrderID)
	if err != nil {
		// Handle error, for example send a 404 Not Found response
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket request not found"})
		return
	}

	// Send a 200 OK response
	c.JSON(http.StatusOK, gin.H{"status": "Ticket request cancelled"})
}

func (trc *TicketOrderController) GuestCancelTicketOrder(c *gin.Context) {
	// Get the ID of the ticket request from the URL parameters
	ugkthid := c.Param("ugkthid")
	if ugkthid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	requestToken := c.Query("request_token")
	if requestToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing request token"})
		return
	}

	ticketOrderID := c.Param("ticketOrderID")

	var user models.User
	if err := trc.Service.DB.
		Preload("TicketOrders").
		Where("id = ? AND request_token = ?", ugkthid, requestToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid request token"})
		return
	}

	userTicketOrder := user.TicketOrders[0]
	if fmt.Sprint(userTicketOrder.ID) != ticketOrderID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ticket request ID"})
		return
	}

	// Use your database or service layer to find the ticket request by ID and cancel it
	err := trc.Service.CancelTicketOrder(ticketOrderID)
	if err != nil {
		// Handle error, for example send a 404 Not Found response
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket request not found"})
		return
	}

	// Delete the user
	if err := trc.Service.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send a 200 OK response
	c.JSON(http.StatusOK, gin.H{"status": "Ticket request cancelled"})
}

func (trc *TicketOrderController) UpdateAddOns(c *gin.Context) {
	// Get the ID of the ticket request from the URL parameters
	ticketOrderIDstring := c.Param("ticketOrderID")
	ticketOrderID, err := strconv.Atoi(ticketOrderIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket request ID"})
		return
	}

	ticketReleaseIDstring := c.Param("ticketReleaseID")
	ticketReleaseID, err := strconv.Atoi(ticketReleaseIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket release ID"})
		return
	}

	// Create a struct to hold the request body
	var request []types.SelectedAddOns

	// Bind the request body to the struct
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use your database or service layer to update the add-ons for the ticket request
	var aerr *types.ErrorResponse = trc.Service.UpdateAddOns(request, ticketOrderID, ticketReleaseID)
	if aerr != nil {
		// Handle aerror, for example send a 404 Not Found response
		c.JSON(aerr.StatusCode, gin.H{"error": aerr.Message})
		return
	}

	// Send a 200 OK response
	c.JSON(http.StatusOK, gin.H{"status": "Add-ons updated"})
}
