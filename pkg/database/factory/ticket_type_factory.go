package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTicketType(eventID uint, name, description string, price float64, quantityTotal uint, isReserved bool, ticketReleaseID uint) *models.TicketType {
	return &models.TicketType{
		EventID:         eventID,
		Name:            name,
		Description:     description,
		Price:           price,
		QuantityTotal:   quantityTotal,
		TicketReleaseID: ticketReleaseID,
	}
}
