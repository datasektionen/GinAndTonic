package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTicketRequest(ticketAmount int, ticketReleaseID, ticketTypeID uint, userUGKthID string, isHandled bool) *models.TicketRequest {
	return &models.TicketRequest{
		TicketAmount:    ticketAmount,
		TicketReleaseID: ticketReleaseID,
		TicketTypeID:    ticketTypeID,
		UserUGKthID:     userUGKthID,
		IsHandled:       isHandled,
	}
}
