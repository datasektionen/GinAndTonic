package surfboard_service_order

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type SurfboardOrderService struct {
	db *gorm.DB
}

func NewSurfboardOrderService(db *gorm.DB) *SurfboardOrderService {
	return &SurfboardOrderService{db: db}
}

func (sos *SurfboardOrderService) CreateOrder(ticketIDs []uint, user *models.User) (*models.Order, error) {

	// Recieves a list of ticket IDs and creates an order for them
	// The order will be created in the database and should return a link to the user to pay for the order
	tx := sos.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var tickets []models.Ticket
	for _, ticketID := range ticketIDs {
		ticket, err := models.GetTicketByID(tx, ticketID)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}

	err := tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return nil, nil
}
