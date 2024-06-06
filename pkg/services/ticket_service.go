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
	// Get all tickets where ticket.ticketOrder.EventID == EventID
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
		Preload("TicketOrder.TicketRelease.Event.Organization").
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
	err := ticket.Delete(ts.DB, "User cancelled ticket")
	if err != nil {
		return &types.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error deleting ticket"}
	}

	// Notify user
	if err := Notify_TicketCancelled(ts.DB,
		&ticket.User,
		&ticket.TicketOrder.TicketRelease.Event.Organization,
		ticket.TicketOrder.TicketRelease.Event.Name); err != nil {
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

	if body.PaymentDeadline != nil && ticket.PaymentDeadline.Valid {
		if !utils.IsEqualTimePtr(&ticket.PaymentDeadline.Time, body.PaymentDeadline) {
			shouldNotifyUser = true
			ticket.PaymentDeadline = sql.NullTime{Time: *body.PaymentDeadline, Valid: true}
		}
	}

	if body.PaymentDeadline != nil {
		ticket.PaymentDeadline = sql.NullTime{Time: *body.PaymentDeadline, Valid: true}
	}

	// Checks the payment deadline, ensure that Event is preloaded
	if err := ticket.ValidatePaymentDeadline(tc.DB); err != nil {
		return nil, err
	}

	if body.CheckedIn != nil {
		ticket.CheckedIn = *body.CheckedIn
	}

	if err := tc.DB.Save(ticket).Error; err != nil {
		return nil, err
	}

	if shouldNotifyUser {
		if err := Notify_UpdatedPaymentDeadlineEmail(tc.DB, int(ticket.ID), &ticket.PaymentDeadline.Time); err != nil {
			return nil, err
		}
	}

	return ticket, nil
}
