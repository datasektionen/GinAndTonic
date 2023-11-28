package services

import (
	"log"
	"net/http"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ErrorResponse struct {
	StatusCode int    // HTTP status code
	Message    string // Error message
}

type TicketRequestService struct {
	DB *gorm.DB
}

func NewTicketRequestService(db *gorm.DB) *TicketRequestService {
	return &TicketRequestService{DB: db}
}

func (trs *TicketRequestService) CreateTicketRequest(ticketRequest *models.TicketRequest) *ErrorResponse {
	if !trs.isTicketReleaseOpen(ticketRequest.TicketReleaseID) {
		log.Println("Ticket release is not open")
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release is not open"}
	}

	if !trs.isTicketTypeValid(ticketRequest.TicketTypeID, ticketRequest.TicketReleaseID) {
		log.Println("Ticket type is not valid for ticket release")
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket type is not valid for ticket release"}
	}

	var ticketRelease models.TicketRelease
	if err := trs.DB.Where("id = ?", ticketRequest.TicketReleaseID).First(&ticketRelease).Error; err != nil {
		log.Println("Error getting ticket release")
		return &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release"}
	}

	var ticketReleaseMethodDetail models.TicketReleaseMethodDetail
	if err := trs.DB.Where("id = ?", ticketRelease.TicketReleaseMethodDetailID).First(&ticketReleaseMethodDetail).Error; err != nil {
		log.Println("Error getting ticket release method detail")
		return &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release method detail"}
	}

	if trs.userAlreadyHasATicketToEvent(ticketRequest.UserUGKthID, ticketRequest.TicketReleaseID, &ticketReleaseMethodDetail, ticketRequest.TicketAmount) {
		log.Println("User cannot request more tickets to this event")
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "User cannot request more tickets to this event"}
	}

	if err := trs.DB.Create(ticketRequest).Error; err != nil {
		log.Println("Error creating ticket request")
		return &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket request"}
	}

	return nil
}

func (trs *TicketRequestService) GetTicketRequests(UGKthID string) ([]models.TicketRequest, *ErrorResponse) {
	var ticketRequests []models.TicketRequest
	if err := trs.DB.Preload("TicketType").Preload("TicketRelease.TicketReleaseMethodDetail").Where("user_ug_kth_id = ?", UGKthID).Find(&ticketRequests).Error; err != nil {
		return nil, &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error listing ticket requests"}
	}
	return ticketRequests, nil
}

// Additional private methods (isTicketReleaseOpen, isTicketTypeValid, userAlreadyHasATicketToEvent) go here...

/*
List all ticket request for the user
*/
func (trs *TicketRequestService) Get(c *gin.Context) {
	var ticketRequests []models.TicketRequest

	UGKthID, _ := c.Get("ugkthid")

	if err := trs.DB.Preload("TicketType").Preload("TicketRelease.TicketReleaseMethodDetail").Where("user_ug_kth_id = ?", UGKthID.(string)).Find(&ticketRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error listing the ticket requests"})
		return
	}

	c.JSON(http.StatusOK, ticketRequests)
}

func (trs *TicketRequestService) isTicketReleaseOpen(ticketReleaseID uint) bool {
	var ticketRelease models.TicketRelease
	now := time.Now().Unix()

	if err := trs.DB.Where("id = ?", ticketReleaseID).First(&ticketRelease).Error; err != nil {
		return false
	}
	return now >= ticketRelease.Open && now <= ticketRelease.Close
}

func (trs *TicketRequestService) isTicketTypeValid(ticketTypeID uint, ticketReleaseID uint) bool {
	var count int64
	trs.DB.Model(&models.TicketRelease{}).Joins("JOIN ticket_types ON ticket_types.ticket_release_id = ticket_releases.id").
		Where("ticket_types.id = ? AND ticket_releases.id = ?", ticketTypeID, ticketReleaseID).Count(&count)

	return count > 0
}

func (trs *TicketRequestService) userAlreadyHasATicketToEvent(userUGKthID string, ticketReleaseID uint, ticketReleaseMethodDetail *models.TicketReleaseMethodDetail, requestedAmount int) bool {
	var ticketRelease models.TicketRelease

	// Get event ID from ticket release
	if err := trs.DB.Where("id = ?", ticketReleaseID).First(&ticketRelease).Error; err != nil {
		return false
	}

	var totalRequestedAmount int64
	trs.DB.Model(&models.TicketRequest{}).
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Joins("JOIN events ON ticket_releases.event_id = events.id").
		Where("ticket_requests.user_ug_kth_id = ? AND events.id = ?", userUGKthID, ticketRelease.EventID).
		Select("SUM(ticket_requests.ticket_amount)").
		Row().Scan(&totalRequestedAmount)

	newRequestedTicketsAmount := int64(ticketReleaseMethodDetail.MaxTicketsPerUser) - int64(requestedAmount)

	return totalRequestedAmount > newRequestedTicketsAmount
}
