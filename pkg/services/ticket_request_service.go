package services

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketRequestService struct {
	DB *gorm.DB
}

func NewTicketRequestService(db *gorm.DB) *TicketRequestService {
	return &TicketRequestService{DB: db}
}

func (trs *TicketRequestService) CreateTicketRequests(ticketRequests []models.TicketRequest) (modelTicketRequests []models.TicketRequest, err *types.ErrorResponse) {
	// Start transaction
	trx := trs.DB.Begin()

	for _, ticketRequest := range ticketRequests {
		tr, err := trs.CreateTicketRequest(trx, &ticketRequest)
		if err != nil {
			trx.Rollback()
			return nil, err
		}
		modelTicketRequests = append(modelTicketRequests, *tr)
	}

	trx.Commit()

	return modelTicketRequests, nil
}

func (trs *TicketRequestService) CreateTicketRequest(
	transaction *gorm.DB,
	ticketRequest *models.TicketRequest,
) (mTicketRequest *models.TicketRequest, err *types.ErrorResponse) {
	var user models.User

	if err := transaction.Where("ug_kth_id = ?", ticketRequest.UserUGKthID).First(&user).Error; err != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting user"}
	}

	// TODO prehaps create log file this?
	var ticketRelease models.TicketRelease
	if err := transaction.Preload("ReservedUsers").Preload("Event.Organization").Where("id = ?", ticketRequest.TicketReleaseID).First(&ticketRelease).Error; err != nil {
		log.Println("Error getting ticket release")
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release"}
	}

	if user.IsExternal && !ticketRelease.AllowExternal {
		log.Println("External user cannot request tickets to this event")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "External user cannot request these tickets"}
	}

	if ticketRelease.HasAllocatedTickets {
		log.Println("Ticket release has allocated tickets")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release has allocated tickets"}
	}

	if !trs.isTicketReleaseOpen(ticketRequest.TicketReleaseID) {
		log.Println("Ticket release is not open")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release is not open"}
	}

	if !trs.isTicketTypeValid(ticketRequest.TicketTypeID, ticketRequest.TicketReleaseID) {
		log.Println("Ticket type is not valid for ticket release")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket type is not valid for ticket release"}
	}

	var ticketReleaseMethodDetail models.TicketReleaseMethodDetail
	if err := transaction.Preload("TicketReleaseMethod").Where("id = ?", ticketRelease.TicketReleaseMethodDetailID).First(&ticketReleaseMethodDetail).Error; err != nil {
		log.Println("Error getting ticket release method detail")
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release method detail"}
	}

	if trs.userAlreadyHasATicketToEvent(&user, &ticketRelease, &ticketReleaseMethodDetail, ticketRequest.TicketAmount) {
		log.Println("User cannot request more tickets to this event")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "User cannot request more tickets to this event"}
	}

	if !trs.checkReservedTicketRelease(&ticketRelease, &user) {
		log.Println("User does not have access to this ticket release")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "You dont have access to this ticket release"}
	}

	mTicketRequest = &models.TicketRequest{
		UserUGKthID:     user.UGKthID,
		TicketTypeID:    ticketRequest.TicketTypeID,
		TicketReleaseID: ticketRelease.ID,
		TicketAmount:    ticketRequest.TicketAmount,
	}

	if err := transaction.Create(mTicketRequest).Error; err != nil {
		log.Println("Error creating ticket request")
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket request"}
	}

	// Check the relead method
	if ticketReleaseMethodDetail.TicketReleaseMethod.MethodName == string(models.RESERVED_TICKET_RELEASE) {
		// We can allocated the ticket to the user directly if there are tickets_available
		// Otherwise fail the request
		var ticketCount int64
		if err := transaction.Model(&models.TicketRequest{}).Where("ticket_release_id = ? AND is_handled = ?", ticketRelease.ID, true).Count(&ticketCount).Error; err != nil {
			transaction.Rollback()
			return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket count"}
		}

		if int64(ticketRelease.TicketsAvailable) < ticketCount+int64(ticketRequest.TicketAmount) {
			transaction.Rollback()
			return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Not enough tickets available"}
		}
	}

	return mTicketRequest, nil
}

func (trs *TicketRequestService) GetTicketRequestsForUser(UGKthID string, ids *[]int) ([]models.TicketRequest, *types.ErrorResponse) {
	ticketRequests, err := models.GetAllValidUsersTicketRequests(trs.DB, UGKthID, ids)

	if err != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket requests"}
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
		return errors.New("ticket request is already allocated to a ticket, cancel the ticket instead")
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

func (trs *TicketRequestService) GetTicketRequest(ticketRequestID int) (ticketRequest *models.TicketRequest, err *types.ErrorResponse) {
	// Use your database or service layer to find the ticket request by ID
	ticketRequest, err2 := models.GetValidTicketReqeust(trs.DB, uint(ticketRequestID))
	if err2 != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket request"}
	}

	return ticketRequest, nil
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

	newRequestedTicketsAmount := int64(ticketReleaseMethodDetail.MaxTicketsPerUser) - int64(requestedAmount)

	return totalRequestedAmount > newRequestedTicketsAmount
}
