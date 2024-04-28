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

func (trs *TicketRequestService) CreateTicketRequests(ticketRequests []models.TicketRequest,
	selectedAddOns *[]types.SelectedAddOns) (modelTicketRequests []models.TicketRequest, err *types.ErrorResponse) {
	// Start transaction
	trx := trs.DB.Begin()

	// Check if all ticket requests are for the same ticket release
	if !allTicketsRequestIsForSameTicketRelease(ticketRequests) {
		trx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "All ticket requests must be for the same ticket release"}
	}

	addonErr := ValidateAddOnsForTicketRequest(trx, *selectedAddOns, int(ticketRequests[0].TicketReleaseID))
	if addonErr != nil {
		trx.Rollback()
		return nil, addonErr
	}

	var ticketRelease models.TicketRelease
	if err := trx.Preload("TicketReleaseMethodDetail").Where("id = ?", ticketRequests[0].TicketReleaseID).First(&ticketRelease).Error; err != nil {
		trx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release"}
	}

	if len(ticketRequests) > int(ticketRelease.TicketReleaseMethodDetail.MaxTicketsPerUser) {
		trx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Too many tickets requested"}
	}

	for _, ticketRequest := range ticketRequests {
		tr, err := trs.CreateTicketRequest(trx, &ticketRequest)
		if err != nil {
			trx.Rollback()
			return nil, err
		}

		// Should be updated to handle multiple ticket requests
		modelTicketRequests = append(modelTicketRequests, *tr)
	}

	for _, selectedAddOn := range *selectedAddOns {
		// TODO: needs to change if multiple ticket requests are allowed in the future
		trId := modelTicketRequests[0].ID

		ticketAddon := models.TicketAddOn{
			TicketRequestID: &trId,
			AddOnID:         uint(selectedAddOn.ID),
			Quantity:        selectedAddOn.Quantity,
		}

		if err := trx.Create(&ticketAddon).Error; err != nil {
			trx.Rollback()
			return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket add-on"}
		}
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

	var ticketRelease models.TicketRelease
	if err := transaction.Preload("ReservedUsers").Preload("Event.Organization").Where("id = ?", ticketRequest.TicketReleaseID).First(&ticketRelease).Error; err != nil {
		log.Println("Error getting ticket release")
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release"}
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

	// Check if the users email is used for any other ticket request
	if user.IsGuestCustomer(transaction) {
		var existingUsersWithSameEmail []models.User
		if err := transaction.Where("email = ?", user.Email).Find(&existingUsersWithSameEmail).Error; err != nil {
			log.Println("Error getting users with same email")
			return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting users with same email"}
		}

		// If any of these users has a ticke to this event we should not allow the request if its over the limit
		for _, existingUser := range existingUsersWithSameEmail {
			if trs.userAlreadyHasATicketToEvent(&existingUser, &ticketRelease, &ticketReleaseMethodDetail, ticketRequest.TicketAmount) {
				return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "User cannot request more tickets to this event"}
			}
		}
	}

	if trs.userAlreadyHasATicketToEvent(&user, &ticketRelease, &ticketReleaseMethodDetail, ticketRequest.TicketAmount) {
		log.Println("User cannot request more tickets to this event")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "User cannot request more tickets to this event"}
	}

	if !trs.checkReservedTicketRelease(&ticketRelease, &user) {
		log.Println("User does not have access to this ticket release")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "You don't have access to this ticket release"}
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

	// Check the release method
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

	// Check if ticket request is allocated to a ticket
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

func (trs *TicketRequestService) UpdateAddOns(selectedAddOns []types.SelectedAddOns, ticketRequestID, ticketReleaseID int) *types.ErrorResponse {
	// Use your database layer to find the ticket xrequest by ID and update the add-ons
	// This is just a placeholder implementation, replace it with your actual code
	// Start by removing all ticket add-ons for the ticket request
	tx := trs.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ticketRequest models.TicketRequest
	if err := tx.Where("id = ?", ticketRequestID).First(&ticketRequest).Error; err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket request"}
	}

	if ticketRequest.IsHandled {
		return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket request has been handled"}
	}

	addonErr := ValidateAddOnsForTicketRequest(tx, selectedAddOns, ticketReleaseID)
	if addonErr != nil {
		tx.Rollback()
		return addonErr
	}

	if err := tx.Unscoped().Where("ticket_request_id = ?", ticketRequestID).Delete(&models.TicketAddOn{}).Error; err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error deleting ticket add-ons"}
	}

	// Then add the new add-ons
	for _, selectedAddOn := range selectedAddOns {
		trID := uint(ticketRequestID)
		ticketAddOn := models.TicketAddOn{
			TicketRequestID: &trID,
			AddOnID:         uint(selectedAddOn.ID),
			Quantity:        selectedAddOn.Quantity,
		}

		if err := tx.Create(&ticketAddOn).Error; err != nil {
			return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket add-on"}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error updating ticket add-ons"}
	}

	return nil
}

func allTicketsRequestIsForSameTicketRelease(ticketRequests []models.TicketRequest) bool {
	if len(ticketRequests) == 0 {
		return false
	}

	ticketReleaseID := ticketRequests[0].TicketReleaseID

	for _, ticketRequest := range ticketRequests {
		if ticketRequest.TicketReleaseID != ticketReleaseID {
			return false
		}
	}

	return true
}
