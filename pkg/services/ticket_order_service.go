package services

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type TicketOrderService struct {
	DB *gorm.DB
}

func NewTicketOrderService(db *gorm.DB) *TicketOrderService {
	return &TicketOrderService{DB: db}
}

func (trs *TicketOrderService) CreateTicketOrder(req models.TicketOrder,
	selectedAddOns *[]types.SelectedAddOns) (*models.TicketOrder, *types.ErrorResponse) {
	// Start transaction
	trx := trs.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	var ticketRelease models.TicketRelease
	if err := trx.Preload("TicketReleaseMethodDetail").Where("id = ?", req.TicketReleaseID).First(&ticketRelease).Error; err != nil {
		trx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release"}
	}

	addonErr := ValidateAddOnsForTicket(trx, *selectedAddOns, int(ticketRelease.ID))
	if addonErr != nil {
		trx.Rollback()
		return nil, addonErr
	}

	if len(req.Tickets) > int(ticketRelease.TicketReleaseMethodDetail.MaxTicketsPerUser) {
		trx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Too many tickets requested"}
	}

	modelTicketOrder := models.TicketOrder{
		TicketReleaseID: req.TicketReleaseID,
		UserUGKthID:     req.UserUGKthID,
		IsHandled:       false,
		Type:            models.TicketOrderRequest,
		NumTickets:      len(req.Tickets),
	}

	if err := trx.Create(&modelTicketOrder).Error; err != nil {
		trx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket order"}
	}

	for _, ticket := range req.Tickets {
		tr, err := trs.CreateticketOrder(trx, &modelTicketOrder, &ticket)
		if err != nil {
			trx.Rollback()
			return nil, err
		}
		for _, selectedAddOn := range *selectedAddOns {
			ticketAddon := models.TicketAddOn{
				TicketID: &tr.ID,
				AddOnID:  uint(selectedAddOn.ID),
				Quantity: selectedAddOn.Quantity,
			}

			if err := trx.Create(&ticketAddon).Error; err != nil {
				trx.Rollback()
				return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket add-on"}
			}
		}
	}

	trx.Commit()

	if trx.Error != nil {
		trx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket order"}
	}

	return &modelTicketOrder, nil
}

func (trs *TicketOrderService) CreateticketOrder(
	transaction *gorm.DB,
	ticketOrder *models.TicketOrder,
	ticket *models.Ticket,
) (mTicket *models.Ticket, err *types.ErrorResponse) {
	var user models.User
	if err := transaction.Where("id = ?", ticketOrder.UserUGKthID).First(&user).Error; err != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting user"}
	}

	var ticketRelease models.TicketRelease
	if err := transaction.Preload("ReservedUsers").Preload("Event.Organization").Where("id = ?", ticketOrder.TicketReleaseID).First(&ticketRelease).Error; err != nil {
		log.Println("Error getting ticket release")
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket release"}
	}

	if ticketRelease.HasAllocatedTickets {
		log.Println("Ticket release has allocated tickets")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release has allocated tickets"}
	}

	if !trs.isTicketReleaseOpen(ticketOrder.TicketReleaseID) {
		log.Println("Ticket release is not open")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket release is not open"}
	}

	if !trs.isTicketTypeValid(ticket.TicketTypeID, ticketOrder.TicketReleaseID) {
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
			if trs.userAlreadyHasATicketToEvent(&existingUser, &ticketRelease, &ticketReleaseMethodDetail, ticketOrder.NumTickets) {
				return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "User cannot request more tickets to this event"}
			}
		}
	}

	if trs.userAlreadyHasATicketToEvent(&user, &ticketRelease, &ticketReleaseMethodDetail, ticketOrder.NumTickets) {
		log.Println("User cannot request more tickets to this event")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "User cannot request more tickets to this event"}
	}

	if !trs.checkReservedTicketRelease(&ticketRelease, &user) {
		log.Println("User does not have access to this ticket release")
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "You don't have access to this ticket release"}
	}

	mTicket = &models.Ticket{
		TicketOrderID: ticketOrder.ID,
		TicketTypeID:  ticket.TicketTypeID,
		UserUGKthID:   user.UGKthID,
	}

	if err := transaction.Create(mTicket).Error; err != nil {
		log.Println("Error creating order ticket")
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket request"}
	}

	// Check the release method
	// TODO MOVE THIS SOMEWHERE ELSE
	if ticketReleaseMethodDetail.TicketReleaseMethod.MethodName == string(models.RESERVED_TICKET_RELEASE) {
		// We can allocated the ticket to the user directly if there are tickets_available
		// Otherwise fail the request
		var ticketCount int64
		if err := transaction.
			Model(&models.Ticket{}).
			Joins("JOIN ticket_orders ON ticket_orders.id = tickets.ticket_order_id").
			Where("ticket_orders.ticket_release_id = ? AND ticket_orders.handled_at IS NOT NULL", ticketRelease.ID).
			Count(&ticketCount).Error; err != nil {
			transaction.Rollback()
			fmt.Println(err.Error())
			return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket count"}
		}

		if int64(ticketRelease.TicketsAvailable) < ticketCount+int64(ticketOrder.NumTickets) {
			transaction.Rollback()
			return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Not enough tickets available"}
		}
	}

	return mTicket, nil
}

func (trs *TicketOrderService) GetTicketOrdersForUser(UGKthID string, ids *[]int) ([]models.TicketOrder, *types.ErrorResponse) {
	ticketOrders, err := models.GetAllValidUsersTicketOrder(trs.DB, UGKthID, ids)
	if err != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket requests"}
	}

	return ticketOrders, nil
}

func (trs *TicketOrderService) CancelTicketOrder(ticketOrderID string) error {
	// Use your database layer to find the ticket request by ID and cancel it
	// This is just a placeholder implementation, replace it with your actual code
	ticketOrder := &models.TicketOrder{}
	result := trs.DB.
		Preload("User").
		Preload("TicketRelease.Event.Organization").Where("id = ?", ticketOrderID).First(ticketOrder)
	if result.Error != nil {
		return result.Error
	}

	user := ticketOrder.User
	org := ticketOrder.TicketRelease.Event.Organization

	// Check if ticket request is allocated to a ticket
	// If the ticket request is allocated to a ticket, it cannot be cancelled
	if ticketOrder.IsHandled {
		return errors.New("ticket order has already been handled")
	}

	if err := trs.DB.Delete(ticketOrder).Error; err != nil {
		return err
	}

	err := Notify_ticketOrderCancelled(trs.DB, &user, &org, ticketOrder.TicketRelease.Event.Name)

	if err != nil {
		return err
	}

	return nil
}

func (trs *TicketOrderService) GetTicketOrder(ticketOrderID int) (ticketOrder *models.TicketOrder, err *types.ErrorResponse) {
	// Use your database or service layer to find the ticket request by ID
	ticketOrder, err2 := models.GetValidTicketOrder(trs.DB, uint(ticketOrderID))
	if err2 != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket request"}
	}

	return ticketOrder, nil
}

func (trs *TicketOrderService) isTicketReleaseOpen(ticketReleaseID uint) bool {
	var ticketRelease models.TicketRelease
	now := time.Now()

	if err := trs.DB.Where("id = ?", ticketReleaseID).First(&ticketRelease).Error; err != nil {
		return false
	}
	return now.After(ticketRelease.Open) && now.Before(ticketRelease.Close)
}

func (trs *TicketOrderService) checkReservedTicketRelease(ticketRelease *models.TicketRelease, user *models.User) bool {
	if !ticketRelease.IsReserved {
		return true
	}

	return ticketRelease.UserHasAccessToTicketRelease(trs.DB, user.UGKthID)
}

func (trs *TicketOrderService) isTicketTypeValid(ticketTypeID uint, ticketReleaseID uint) bool {
	var count int64
	trs.DB.Model(&models.TicketRelease{}).Joins("JOIN ticket_types ON ticket_types.ticket_release_id = ticket_releases.id").
		Where("ticket_types.id = ? AND ticket_releases.id = ?", ticketTypeID, ticketReleaseID).Count(&count)

	return count > 0
}

func (trs *TicketOrderService) userAlreadyHasATicketToEvent(user *models.User,
	ticketRelease *models.TicketRelease,
	ticketReleaseMethodDetail *models.TicketReleaseMethodDetail, requestedAmount int) bool {
	var totalOrderedAmount int64
	trs.DB.Model(&models.TicketOrder{}).
		Joins("JOIN ticket_releases ON ticket_orders.ticket_release_id = ticket_releases.id").
		Joins("JOIN events ON ticket_releases.event_id = events.id").
		Where("ticket_orders.user_ug_kth_id = ? AND ticket_releases.id = ?", user.UGKthID, ticketRelease.ID).
		Select("SUM(ticket_orders.num_tickets)").
		Row().Scan(&totalOrderedAmount)

	newRequestedTicketsAmount := int64(ticketReleaseMethodDetail.MaxTicketsPerUser) - int64(requestedAmount)

	return totalOrderedAmount > newRequestedTicketsAmount
}

func (trs *TicketOrderService) UpdateAddOns(selectedAddOns []types.SelectedAddOns, ticketOrderID, ticketReleaseID int) *types.ErrorResponse {
	// Use your database layer to find the ticket xrequest by ID and update the add-ons
	// This is just a placeholder implementation, replace it with your actual code
	// Start by removing all ticket add-ons for the ticket request
	tx := trs.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ticketOrder models.TicketOrder
	if err := tx.Preload("Tickets").Where("id = ?", ticketOrderID).First(&ticketOrder).Error; err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket request"}
	}

	if ticketOrder.IsHandled {
		return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket request has been handled"}
	}

	addonErr := ValidateAddOnsForTicket(tx, selectedAddOns, ticketReleaseID)
	if addonErr != nil {
		tx.Rollback()
		return addonErr
	}

	if err := tx.Unscoped().Where("ticket_request_id = ?", ticketOrderID).Delete(&models.TicketAddOn{}).Error; err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error deleting ticket add-ons"}
	}

	// Then add the new add-ons
	for _, ticket := range ticketOrder.Tickets {
		trID := ticket.ID

		for _, selectedAddOn := range selectedAddOns {
			ticketAddOn := models.TicketAddOn{
				TicketID: &trID,
				AddOnID:  uint(selectedAddOn.ID),
				Quantity: selectedAddOn.Quantity,
			}

			if err := tx.Create(&ticketAddOn).Error; err != nil {
				return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error creating ticket add-on"}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error updating ticket add-ons"}
	}

	return nil
}
