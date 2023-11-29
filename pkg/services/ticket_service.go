package services

import (
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
