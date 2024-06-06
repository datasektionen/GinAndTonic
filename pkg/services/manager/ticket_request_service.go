package manager_service

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type ManagerTicketOrderService struct {
	DB *gorm.DB
}

func NewManagerTicketOrderService(db *gorm.DB) *ManagerTicketOrderService {
	return &ManagerTicketOrderService{DB: db}
}

// DeleteticketOrder is a method that deletes a ticket request.
func (trc ManagerTicketOrderService) DeleteTicketOrder(ticketOrderIds []int) *types.ErrorResponse {
	var ticketOrders []models.TicketOrder
	err := trc.DB.Where("id IN (?)", ticketOrderIds).Find(&ticketOrders).Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Ticket request not found"}
	}

	tx := trc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ticketOrder := range ticketOrders {
		if ticketOrder.IsHandled {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Ticket request %d has been handled", ticketOrder.ID)}
		}

		err := ticketOrder.Delete(tx, "Deleted by manager")

		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to delete ticket request %d", ticketOrder.ID)}
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Failed to delete ticket requests"}
	}

	return nil
}

// UndeleteticketOrder is a method that undeletes a ticket request.
func (trc ManagerTicketOrderService) UndeleteTicketOrders(ticketOrderIds []int) *types.ErrorResponse {
	var ticketOrders []models.TicketOrder
	err := trc.DB.Unscoped().Where("id IN (?) AND deleted_at IS NOT NULL", ticketOrderIds).Find(&ticketOrders).Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Ticket request not found"}
	}

	tx := trc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ticketOrder := range ticketOrders {
		if ticketOrder.IsHandled {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Ticket request %d has been handled", ticketOrder.ID)}
		}

		err = tx.Unscoped().Model(&ticketOrder).Update("deleted_at", nil).Error
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to undelete ticket request %d", ticketOrder.ID)}
		}

		err = tx.Model(&ticketOrder).Unscoped().UpdateColumn("deleted_reason", "").Error
		if err != nil {
			tx.Rollback()
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Failed to undelete ticket request %d", ticketOrder.ID)}
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: "Failed to undelete ticket requests"}
	}

	return nil
}
