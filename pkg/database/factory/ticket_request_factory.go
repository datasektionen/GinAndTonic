package factory

import (
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func NewTicketRequest(ticketAmount int, ticketReleaseID, ticketTypeID uint, userUGKthID string, isHandled bool, createdAt time.Time) *models.TicketRequest {
	return &models.TicketRequest{
		TicketAmount:    ticketAmount,
		TicketReleaseID: ticketReleaseID,
		TicketTypeID:    ticketTypeID,
		UserUGKthID:     userUGKthID,
		IsHandled:       isHandled,
		Model: gorm.Model{
			CreatedAt: createdAt,
		},
	}
}
