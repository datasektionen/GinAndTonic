package services

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

type TicketService struct {
	DB *gorm.DB
}

func NewTicketService(db *gorm.DB) *TicketService {
	return &TicketService{DB: db}
}

func (ts *TicketService) GetAllTicketsToEvent(eventID int) (tickets []models.Ticket, err error) {
	// Get all tickets where ticket.TicketRequest.EventID == EventID
	tickets, err = models.GetTicketsToEvent(ts.DB, uint(eventID))
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (ts *TicketService) GetTicketToEvent(eventID, ticketID int) (ticket models.Ticket, err error) {
	// Get ticket where ticket.ID == ticketID
	ticket, err = models.GetTicketToEvent(ts.DB, uint(eventID), uint(ticketID))

	if err != nil {
		return models.Ticket{}, err
	}

	return ticket, nil
}

func (ts *TicketService) GetTicketForUser(UGKthID string) ([]models.Ticket, *types.ErrorResponse) {
	tickets, err := models.GetAllValidUsersTicket(ts.DB, UGKthID)

	if err != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket requests"}
	}

	return tickets, nil
}

func (ts *TicketService) CancelTicket(ugKthID string, ticketID int) *types.ErrorResponse {
	// Get ticket
	var ticket models.Ticket
	if err := ts.DB.
		Preload("User").
		Preload("TicketRequest.TicketRelease.Event.Organization").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket"}
	}

	// Check if user is owner of ticket
	if ticket.User.UGKthID != ugKthID {
		return &types.ErrorResponse{StatusCode: http.StatusForbidden, Message: "You are not the owner of this ticket"}
	}

	// Check if ticket is already refunded
	if ticket.Refunded {
		return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket is already refunded"}
	}

	// Check if ticket is paid
	if ticket.IsPaid {
		return &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket is already paid"}
	}

	// Delete ticket
	err := ticket.Delete(ts.DB)
	if err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error deleting ticket"}
	}

	// Notify user
	if err := Notify_TicketCancelled(ts.DB, &ticket.User, &ticket.TicketRequest.TicketRelease.Event.Organization, ticket.TicketRequest.TicketRelease.Event.Name); err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error notifying user"}
	}

	return nil
}

func (ts *TicketService) CheckInViaQrCode(qrCode string) (ticket *models.Ticket, err *types.ErrorResponse) {
	// Get ticket
	if err := ts.DB.
		Preload("User").
		Where("qr_code = ?", qrCode).First(&ticket).Error; err != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket"}
	}

	if ticket.QrCode != qrCode {
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Invalid QR code"}
	}

	if ticket.CheckedIn {
		return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket is already checked in"}
	}

	// Check in ticket
	ticket.CheckedIn = true
	ticket.CheckedInAt = sql.NullTime{Time: time.Now(), Valid: true}

	// Save ticket
	if err := ts.DB.Save(&ticket).Error; err != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error saving ticket"}
	}

	return ticket, nil
}

func (tc *TicketService) UpdateTicket(ticket *models.Ticket, body *types.UpdateTicketBody) (*models.Ticket, error) {
	// check if body.PaymentDeadline is set and different from ticket.PaymentDeadline
	shouldNotifyUser := false

	if body.PaymentDeadline != nil {
		if !utils.IsEqualTimePtr(ticket.PaymentDeadline, body.PaymentDeadline) {
			shouldNotifyUser = true
			ticket.PaymentDeadline = body.PaymentDeadline
		}
	}

	if body.PaymentDeadline != nil {
		ticket.PaymentDeadline = body.PaymentDeadline
	}

	// Checks the payment deadline, ensure that Event is preloaded
	if err := ticket.ValidatePaymentDeadline(); err != nil {
		return nil, err
	}

	if body.CheckedIn != nil {
		ticket.CheckedIn = *body.CheckedIn
	}

	if err := tc.DB.Save(ticket).Error; err != nil {
		return nil, err
	}

	if shouldNotifyUser {
		if err := Notify_UpdatedPaymentDeadlineEmail(tc.DB, int(ticket.ID), ticket.PaymentDeadline); err != nil {
			return nil, err
		}
	}

	return ticket, nil
}

func (tc *TicketService) UpdateTicketType(ticketRequestID int, body *types.UpdateTicketTypeBody) (*models.TicketRequest, *types.ErrorResponse) {
	// Start a new transaction
	tx := tc.DB.Begin()
	if tx.Error != nil {
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error starting transaction"}
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get ticket request
	var ticketRequest models.TicketRequest
	if err := tx.
		Preload("TicketRelease").
		Preload("Tickets").
		Where("id = ?", ticketRequestID).First(&ticketRequest).Error; err != nil {
		tx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket request"}
	}

	var ticketType models.TicketType
	if err := tx.
		Where("id = ?", body.TicketTypeID).First(&ticketType).Error; err != nil {
		tx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket type"}
	}

	if ticketRequest.IsHandled {
		// Change it of the ticket instead

		ticket := ticketRequest.Tickets[0]

		if ticket.IsPaid {
			tx.Rollback()
			return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket is already paid"}
		}

		if ticket.Refunded {
			tx.Rollback()
			return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket is already refunded"}
		}

		if ticket.CheckedIn {
			tx.Rollback()
			return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket is already checked in"}
		}

		// If these pass we can update the ticket type
		ticketRequest.TicketTypeID = ticketType.ID

		// Remove any pending transactions
		var transaction models.Transaction
		if err := tx.
			Where("ticket_id = ?", ticket.ID).First(&transaction).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				tx.Rollback()
				return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting transaction"}
			}
		}

		if ticketType.ID != 0 && ticketType.Price == 0 {
			// Allocate free ticket
			ticket.IsPaid = true

			if err := tx.Save(&ticket).Error; err != nil {
				tx.Rollback()
				return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error saving ticket"}
			}
		}

		if transaction.ID != 0 {
			if transaction.Status == models.TransactionStatusCompleted {
				tx.Rollback()
				return nil, &types.ErrorResponse{StatusCode: http.StatusBadRequest, Message: "A completed transaction exists for this ticket"}
			}

			if err := tx.Delete(&transaction).Error; err != nil {
				tx.Rollback()
				return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error deleting transaction"}
			}
		}
	} else {
		ticketRequest.TicketTypeID = ticketType.ID
	}

	if err := tx.Save(&ticketRequest).Error; err != nil {
		tx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error saving ticket request"}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error committing transaction"}
	}

	return &ticketRequest, nil
}
