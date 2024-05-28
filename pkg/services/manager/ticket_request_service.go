package manager_service

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type ManagerTicketRequestService struct {
	DB *gorm.DB
}

func NewManagerTicketRequestService(db *gorm.DB) *ManagerTicketRequestService {
	return &ManagerTicketRequestService{DB: db}
}

// DeleteTicketRequest is a method that deletes a ticket request.
func (trc *ManagerTicketRequestService) DeleteTicketRequests(ticketRequestIDs []int) *types.ErrorResponse {
	var ticketRequests []models.TicketRequest
	err := trc.DB.Where("id IN (?)", ticketRequestIDs).Find(&ticketRequests).Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Ticket request not found"}
	}

	tx := trc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ticketRequest := range ticketRequests {
		if ticketRequest.IsHandled {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Ticket request %d has been handled", ticketRequest.ID)}
		}

		err := ticketRequest.Delete(tx, "Deleted by manager")

		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to delete ticket request %d", ticketRequest.ID)}
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Failed to delete ticket requests"}
	}

	return nil
}

// UndeleteTicketRequest is a method that undeletes a ticket request.
func (trc *ManagerTicketRequestService) UndeleteTicketRequests(ticketRequestIDs []int) *types.ErrorResponse {
	var ticketRequests []models.TicketRequest
	err := trc.DB.Unscoped().Where("id IN (?) AND deleted_at IS NOT NULL", ticketRequestIDs).Find(&ticketRequests).Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Ticket request not found"}
	}

	tx := trc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ticketRequest := range ticketRequests {
		if ticketRequest.IsHandled {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Ticket request %d has been handled", ticketRequest.ID)}
		}

		err = tx.Unscoped().Model(&ticketRequest).Update("deleted_at", nil).Error
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to undelete ticket request %d", ticketRequest.ID)}
		}

		err = tx.Model(&ticketRequest).Unscoped().UpdateColumn("deleted_reason", "").Error
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to undelete ticket request %d", ticketRequest.ID)}
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Failed to undelete ticket requests"}
	}

	return nil
}
