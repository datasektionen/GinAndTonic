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

func (trs *TicketRequestService) CreateTicketRequests(ticketRequests []models.TicketRequest) *ErrorResponse {
	// Start transaction
	trx := trs.DB.Begin()

	for _, ticketRequest := range ticketRequests {
		err := trs.CreateTicketRequest(&ticketRequest, trx)
		if err != nil {
			trx.Rollback()
			return err
		}
	}

	trx.Commit()
	return nil
}

func (trs *TicketRequestService) CreateTicketRequest(ticketRequest *models.TicketRequest, transaction *gorm.DB) *ErrorResponse {
	var user models.User

	if err := transaction.Where("ug_kth_id = ?", ticketRequest.UserUGKthID).First(&user).Error; err != nil {
		return &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting user"}
	}

	var ticketRelease models.TicketRelease
	if err := transaction.Preload("ReservedUsers").Where("id = ?", ticketRequest.TicketReleaseID).First(&ticketRelease).Error; err != nil {
		log.Println("Error getting ticket release")
		return &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release"}
	}

	if ticketRelease.HasAllocatedTickets {
		log.Println("Ticket release has allocated tickets")
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release has allocated tickets"}
	}

	if !trs.isTicketReleaseOpen(ticketRequest.TicketReleaseID) {
		log.Println("Ticket release is not open")
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release is not open"}
	}

	if !trs.isTicketTypeValid(ticketRequest.TicketTypeID, ticketRequest.TicketReleaseID) {
		log.Println("Ticket type is not valid for ticket release")
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket type is not valid for ticket release"}
	}

	var ticketReleaseMethodDetail models.TicketReleaseMethodDetail
	if err := transaction.Where("id = ?", ticketRelease.TicketReleaseMethodDetailID).First(&ticketReleaseMethodDetail).Error; err != nil {
		log.Println("Error getting ticket release method detail")
		return &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release method detail"}
	}

	if trs.userAlreadyHasATicketToEvent(&user, &ticketRelease, &ticketReleaseMethodDetail, ticketRequest.TicketAmount) {
		log.Println("User cannot request more tickets to this event")
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "User cannot request more tickets to this event"}
	}

	if !trs.checkReservedTicketRelease(&ticketRelease, &user) {
		log.Println("User does not have access to this ticket release")
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "You dont have access to this ticket release"}
	}

	created_ticket_request := models.TicketRequest{
		UserUGKthID:     user.UGKthID,
		TicketTypeID:    ticketRequest.TicketTypeID,
		TicketReleaseID: ticketRelease.ID,
		TicketAmount:    ticketRequest.TicketAmount,
	}

	if err := transaction.Create(&created_ticket_request).Error; err != nil {
		log.Println("Error creating ticket request")
		return &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket request"}
	}

	return nil
}

func (trs *TicketRequestService) GetTicketRequestsForUser(UGKthID string) ([]models.TicketRequest, *ErrorResponse) {
	ticketRequests, err := models.GetAllValidUsersTicketRequests(trs.DB, UGKthID)

	if err != nil {
		return nil, &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket requests"}
	}

	return ticketRequests, nil
}

func (trs *TicketRequestService) CancelTicketRequest(ticketRequestID string) error {
	// Use your database layer to find the ticket request by ID and cancel it
	// This is just a placeholder implementation, replace it with your actual code
	ticketRequest := &models.TicketRequest{}
	result := trs.DB.Where("id = ?", ticketRequestID).First(ticketRequest)
	if result.Error != nil {
		return result.Error
	}

	if err := trs.DB.Delete(ticketRequest).Error; err != nil {
		return err
	}

	return nil
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

func (trs *TicketRequestService) checkReservedTicketRelease(ticketRelease *models.TicketRelease, user *models.User) bool {
	if !ticketRelease.IsReserved {
		return true
	}

	return ticketRelease.UserHasAccessToTicketRelease(user)
}

func (trs *TicketRequestService) isTicketTypeValid(ticketTypeID uint, ticketReleaseID uint) bool {
	var count int64
	trs.DB.Model(&models.TicketRelease{}).Joins("JOIN ticket_types ON ticket_types.ticket_release_id = ticket_releases.id").
		Where("ticket_types.id = ? AND ticket_releases.id = ?", ticketTypeID, ticketReleaseID).Count(&count)

	return count > 0
}

func (trs *TicketRequestService) userAlreadyHasATicketToEvent(user *models.User, ticketRelease *models.TicketRelease, ticketReleaseMethodDetail *models.TicketReleaseMethodDetail, requestedAmount int) bool {
	var totalRequestedAmount int64
	trs.DB.Model(&models.TicketRequest{}).
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Joins("JOIN events ON ticket_releases.event_id = events.id").
		Where("ticket_requests.user_ug_kth_id = ? AND events.id = ?", user.UGKthID, ticketRelease.EventID).
		Select("SUM(ticket_requests.ticket_amount)").
		Row().Scan(&totalRequestedAmount)

	newRequestedTicketsAmount := int64(ticketReleaseMethodDetail.MaxTicketsPerUser) - int64(requestedAmount)

	return totalRequestedAmount > newRequestedTicketsAmount
}
