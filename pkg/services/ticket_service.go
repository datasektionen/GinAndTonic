package services

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
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

func (ts *TicketService) EditTicket(eventID, ticketID int, updatedTicket models.Ticket) (ticket models.Ticket, err error) {
	// Get ticket where ticket.ID == ticketID
	ticket, err = models.GetTicketToEvent(ts.DB, uint(eventID), uint(ticketID))
	if err != nil {
		return models.Ticket{}, err
	}

	// Update ticket
	ticket.IsPaid = updatedTicket.IsPaid
	ticket.IsReserve = updatedTicket.IsReserve
	ticket.Refunded = updatedTicket.Refunded

	// Save ticket
	if err := ts.DB.Save(&ticket).Error; err != nil {
		return models.Ticket{}, err
	}

	return ticket, nil
}

func (ts *TicketService) GetTicketForUser(UGKthID string) ([]models.Ticket, *ErrorResponse) {
	tickets, err := models.GetAllValidUsersTicket(ts.DB, UGKthID)

	if err != nil {
		return nil, &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket requests"}
	}

	return tickets, nil
}

func (ts *TicketService) CancelTicket(ugKthID string, ticketID int) *ErrorResponse {
	// Get ticket
	var ticket models.Ticket
	if err := ts.DB.
		Preload("User").
		Preload("TicketRequest.TicketRelease.Event.Organization").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		return &ErrorResponse{StatusCode: http.StatusInternalServerError, Message: "Error getting ticket"}
	}

	// Check if user is owner of ticket
	if ticket.User.UGKthID != ugKthID {
		return &ErrorResponse{StatusCode: http.StatusForbidden, Message: "You are not the owner of this ticket"}
	}

	// Check if ticket is already refunded
	if ticket.Refunded {
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket is already refunded"}
	}

	// Check if ticket is paid
	if ticket.IsPaid {
		return &ErrorResponse{StatusCode: http.StatusBadRequest, Message: "Ticket is already paid"}
	}

	return nil
}
