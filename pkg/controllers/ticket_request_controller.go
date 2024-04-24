package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
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
	UGKthId, exists := c.Get("ugkthid")
	if !exists {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	idsString := c.Query("ids")
	var idsInt []int
	if idsString != "" {
		ids := strings.Split(idsString, ",")
		for _, id := range ids {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket request ID"})
				return
			}
			idsInt = append(idsInt, idInt)
		}
	}

	ticketRequests, err := trc.Service.GetTicketRequestsForUser(UGKthId.(string), &idsInt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ticket_requests": ticketRequests})
}

type TicketRequestCreateRequest struct {
	TicketRequests []models.TicketRequest `json:"ticket_requests"`
	SelectedAddOns []types.SelectedAddOns `json:"selected_add_ons"`
}

// Create a ticket request
func (trc *TicketRequestController) Create(c *gin.Context) {
	var request TicketRequestCreateRequest
	var ticketRequestsIds []int

	UGKthID, _ := c.Get("ugkthid")

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ticketRequests []models.TicketRequest = request.TicketRequests
	// var addOns []models.AddOn = request.AddOns

	for i := range ticketRequests {
		ticketRequests[i].UserUGKthID = UGKthID.(string)
	}

	mTicketRequests, err := trc.Service.CreateTicketRequests(ticketRequests, &request.SelectedAddOns)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	for _, ticketRequest := range mTicketRequests {
		ticketRequestsIds = append(ticketRequestsIds, int(ticketRequest.ID))
	}

	services.Notify_TicketRequestCreated(trc.Service.DB, ticketRequestsIds)

	c.JSON(http.StatusCreated, mTicketRequests)
}

type TicketRequestGuestCreateRequest struct {
	TicketRequests []models.TicketRequest `json:"ticket_requests"`
	SelectedAddOns []types.SelectedAddOns `json:"selected_add_ons"`
	RequestToken   string                 `json:"request_token" binding:"required"`
}

// Guest reqeust controller
func (trc *TicketRequestController) GuestCreate(c *gin.Context) {
	var request TicketRequestGuestCreateRequest
	var ticketRequestsIds []int

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
	if err := trc.Service.DB.Where("ug_kth_id = ? AND request_token = ?", userUGKTHId, request.RequestToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid request token"})
		return
	}

	var ticketRequests []models.TicketRequest = request.TicketRequests

	for i := range ticketRequests {
		ticketRequests[i].UserUGKthID = userUGKTHId
	}

	mTicketRequests, err := trc.Service.CreateTicketRequests(ticketRequests, &request.SelectedAddOns)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	for _, ticketRequest := range mTicketRequests {
		ticketRequestsIds = append(ticketRequestsIds, int(ticketRequest.ID))
	}

	services.Notify_TicketRequestCreated(trc.Service.DB, ticketRequestsIds)

	c.JSON(http.StatusCreated, mTicketRequests)
}

func (trc *TicketRequestController) Get(c *gin.Context) {
	UGKthID, _ := c.Get("ugkthid")
	ticketRequests, err := trc.Service.GetTicketRequestsForUser(UGKthID.(string), nil)

	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusOK, ticketRequests)
}

func (trc *TicketRequestController) CancelTicketRequest(c *gin.Context) {
	// Get the ID of the ticket request from the URL parameters
	ticketRequestID := c.Param("ticketRequestID")

	// Use your database or service layer to find the ticket request by ID and cancel it
	err := trc.Service.CancelTicketRequest(ticketRequestID)
	if err != nil {
		// Handle error, for example send a 404 Not Found response
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket request not found"})
		return
	}

	// Send a 200 OK response
	c.JSON(http.StatusOK, gin.H{"status": "Ticket request cancelled"})
}

func (trc *TicketRequestController) UpdateAddOns(c *gin.Context) {
	// Get the ID of the ticket request from the URL parameters
	ticketRequestIDstring := c.Param("ticketRequestID")
	ticketRequestID, err := strconv.Atoi(ticketRequestIDstring)
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
	var aerr *types.ErrorResponse = trc.Service.UpdateAddOns(request, ticketRequestID, ticketReleaseID)
	if aerr != nil {
		// Handle aerror, for example send a 404 Not Found response
		c.JSON(aerr.StatusCode, gin.H{"error": aerr.Message})
		return
	}

	// Send a 200 OK response
	c.JSON(http.StatusOK, gin.H{"status": "Add-ons updated"})
}
