package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTicket(ticketRequestID uint, isPaid, isReserve bool, userUGKthID string) *models.Ticket {
	return &models.Ticket{
		TicketRequestID: ticketRequestID,
		IsPaid:          isPaid,
		IsReserve:       isReserve,
		UserUGKthID:     userUGKthID,
	}
}
