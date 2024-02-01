package services

import (
	"errors"
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

// Error implements error.
func (*ErrorResponse) Error() string {
	panic("unimplemented")
}

type TicketRequestService struct {
	DB *gorm.DB
}

func NewTicketRequestService(db *gorm.DB) *TicketRequestService {
	return &TicketRequestService{DB: db}
}

func (trs *TicketRequestService) CreateTicketRequests(ticketRequests []models.TicketRequest) (modelTicketRequests []models.TicketRequest, err *ErrorResponse) {
	// Start transaction
	trx := trs.DB.Begin()

	for _, ticketRequest := range ticketRequests {
		tr, err := trs.CreateTicketRequest(&ticketRequest, trx)
		if err != nil {
			trx.Rollback()
			return nil, err
		}
		modelTicketRequests = append(modelTicketRequests, *tr)
	}

	trx.Commit()
	return modelTicketRequests, nil
}

func (trs *TicketRequestService) CreateTicketRequest(ticketRequest *models.TicketRequest, transaction *gorm.DB) (mTicketRequest *models.TicketRequest, err *ErrorResponse) {
	var user models.User

	if err := transaction.Where("ug_kth_id = ?", ticketRequest.UserUGKthID).First(&user).Error; err != nil {
		return nil, &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting user"}
	}

	var ticketRelease models.TicketRelease
	if err := transaction.Preload("ReservedUsers").Where("id = ?", ticketRequest.TicketReleaseID).First(&ticketRelease).Error; err != nil {
		log.Println("Error getting ticket release")
		return nil, &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release"}
	}

	if ticketRelease.HasAllocatedTickets {
		log.Println("Ticket release has allocated tickets")
		return nil, &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release has allocated tickets"}
	}

	if !trs.isTicketReleaseOpen(ticketRequest.TicketReleaseID) {
		log.Println("Ticket release is not open")
		return nil, &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release is not open"}
	}

	if !trs.isTicketTypeValid(ticketRequest.TicketTypeID, ticketRequest.TicketReleaseID) {
		log.Println("Ticket type is not valid for ticket release")
		return nil, &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket type is not valid for ticket release"}
	}

	var ticketReleaseMethodDetail models.TicketReleaseMethodDetail
	if err := transaction.Where("id = ?", ticketRelease.TicketReleaseMethodDetailID).First(&ticketReleaseMethodDetail).Error; err != nil {
		log.Println("Error getting ticket release method detail")
		return nil, &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release method detail"}
	}

	if trs.userAlreadyHasATicketToEvent(&user, &ticketRelease, &ticketReleaseMethodDetail, ticketRequest.TicketAmount) {
		log.Println("User cannot request more tickets to this event")
		return nil, &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "User cannot request more tickets to this event"}
	}

	if !trs.checkReservedTicketRelease(&ticketRelease, &user) {
		log.Println("User does not have access to this ticket release")
		return nil, &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "You dont have access to this ticket release"}
	}

	mTicketRequest = &models.TicketRequest{
		UserUGKthID:     user.UGKthID,
		TicketTypeID:    ticketRequest.TicketTypeID,
		TicketReleaseID: ticketRelease.ID,
		TicketAmount:    ticketRequest.TicketAmount,
	}

	if err := transaction.Create(mTicketRequest).Error; err != nil {
		log.Println("Error creating ticket request")
		return nil, &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket request"}
	}

	return mTicketRequest, nil
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
	result := trs.DB.Preload("User").Preload("TicketRelease.Event.Organization").Where("id = ?", ticketRequestID).First(ticketRequest)
	if result.Error != nil {
		return result.Error
	}

	user := ticketRequest.User
	org := ticketRequest.TicketRelease.Event.Organization

	// Check if ticket request is allocted to a ticket
	// If the ticket request is allocated to a ticket, it cannot be cancelled
	if len(ticketRequest.Tickets) > 0 {
		return errors.New("Ticket request is already allocated to a ticket, cancel the ticket instead")
	}

	if err := trs.DB.Delete(ticketRequest).Error; err != nil {
		return err
	}

	err := Notify_TicketRequestCancelled(trs.DB, &user, &org, ticketRequest.TicketRelease.Event.Name)

	if err != nil {
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

	return ticketRelease.UserHasAccessToTicketRelease(trs.DB, user.UGKthID)
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
		Where("ticket_requests.user_ug_kth_id = ? AND ticket_releases.id = ?", user.UGKthID, ticketRelease.ID).
		Select("SUM(ticket_requests.ticket_amount)").
		Row().Scan(&totalRequestedAmount)

	println(totalRequestedAmount, requestedAmount, ticketReleaseMethodDetail.MaxTicketsPerUser)

	newRequestedTicketsAmount := int64(ticketReleaseMethodDetail.MaxTicketsPerUser) - int64(requestedAmount)

	println(newRequestedTicketsAmount)

	return totalRequestedAmount > newRequestedTicketsAmount
}
