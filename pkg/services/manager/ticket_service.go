package manager_service

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type ManagerTicketService struct {
	DB *gorm.DB
}

func NewManagerTicketService(db *gorm.DB) *ManagerTicketService {
	return &ManagerTicketService{DB: db}
}

// DeleteTicket is a method that deletes a ticket.
func (ts *ManagerTicketService) DeleteTickets(ticketIDs []int) *types.ErrorResponse {
	var ticketRequests []models.Ticket
	err := ts.DB.Preload("TicketRequest").Where("id IN (?)", ticketIDs).Find(&ticketRequests).Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Ticket not found"}
	}

	tx := ts.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ticket := range ticketRequests {
		if !ticket.TicketRequest.IsHandled {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Ticket %d has not been handled", ticket.ID)}
		}

		err = tx.Delete(&ticket).Error
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to delete ticket %d", ticket.ID)}
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Failed to delete tickets"}
	}

	return nil
}

// UndeleteTicket is a method that undeletes a ticket.
func (ts *ManagerTicketService) UndeleteTickets(ticketIDs []int) *types.ErrorResponse {
	var tickets []models.Ticket
	err := ts.DB.Unscoped().Preload("TicketRequest").Where("id IN (?)", ticketIDs).Find(&tickets).Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Ticket not found"}
	}

	tx := ts.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ticket := range tickets {
		if !ticket.TicketRequest.IsHandled {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Ticket %d has not been handled", ticket.ID)}
		}

		err = tx.Model(&ticket).Unscoped().UpdateColumn("deleted_at", nil).Error
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to undelete ticket %d", ticket.ID)}
		}

		err = tx.Model(&ticket.TicketRequest).Unscoped().UpdateColumn("deleted_at", nil).Error
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to undelete ticket request %d", ticket.TicketRequest.ID)}
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Failed to undelete tickets"}
	}

	return nil
}