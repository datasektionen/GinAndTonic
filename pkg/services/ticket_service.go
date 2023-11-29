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

func (ts *TicketService) GetAllTickets(EventID int) (tickets []models.Ticket, err error) {
	if err := ts.DB.Where("event_id = ?", EventID).Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}
