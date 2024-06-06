package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type TicketOrderType string

const (
	TicketOrderRequest TicketOrderType = "request"
	TicketOrderTicket  TicketOrderType = "ticket"
)

type TicketOrder struct {
	gorm.Model
	UserUGKthID     string          `json:"user_ug_kth_id"`
	User            User            `json:"user"`
	TicketReleaseID uint            `json:"ticket_release_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	TicketRelease   TicketRelease   `json:"ticket_release"`
	IsHandled       bool            `json:"is_handled" gorm:"default:false"`
	Tickets         []Ticket        `gorm:"foreignKey:TicketOrderID" json:"requests"`
	NumTickets      int             `json:"num_tickets"`
	TotalAmount     float64         `json:"total_amount"`
	IsPaid          bool            `json:"is_paid" gorm:"default:false"`
	PaidAt          *time.Time      `json:"paid_at" gorm:"default:null"`
	Type            TicketOrderType `json:"type" gorm:"default:'request'"`
	HandledAt       sql.NullTime    `json:"handled_at" gorm:"default:null"`

	DeletedReason string `json:"deleted_reason" gorm:"default:null"`

	OrderID *string `json:"order_id" gorm:"default:null"`
	Order   Order   `json:"order"`
}

func (to *TicketOrder) BeforeSave(tx *gorm.DB) (err error) {
	if to.IsHandled && to.HandledAt.Valid {
		now := time.Now()
		to.HandledAt = sql.NullTime{Time: now, Valid: true}
	}

	return
}

func (to *TicketOrder) Delete(db *gorm.DB, reason string) error {
	// Delete the associated ticketOrder
	if err := db.Model(to).Update("deleted_reason", reason).Error; err != nil {
		return err
	}

	return db.Delete(to).Error
}

// Function to be called on delete
func (to *TicketOrder) BeforeDelete(tx *gorm.DB) (err error) {
	// Delete all tickets associated with the ticket order

	for _, ticket := range to.Tickets {
		err := ticket.Delete(tx, "Ticket order deleted")
		if err != nil {
			return err
		}
	}

	return nil
}

func (to *TicketOrder) IsticketOrder() bool {
	return to.Type == TicketOrderRequest
}

func GetAllUnhandledTicketsByTicketReleaseID(db *gorm.DB, ticketReleaseID uint) ([]Ticket, error) {
	var tickets []Ticket
	if err := db.
		Preload("TicketType").
		Preload("TicketAddOns.AddOn").
		Joins("JOIN ticket_orders ON ticket_orders.ticket_id = tickets.id").
		Preload("TicketOrder.TicketRelease.Event").
		Preload("TicketOrder.TicketRelease.TicketReleaseMethodDetail").
		Preload("TicketOrder.TicketRelease.PaymentDeadline").
		Where("ticket_orders.ticket_release_id = ? AND ticket_orders.is_handled = ?", ticketReleaseID, false).
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetAllValidUsersTicketOrder(db *gorm.DB, userUGKthID string, ids *[]int) ([]TicketOrder, error) {
	var ticketOrder []TicketOrder

	query := db.
		Unscoped().
		Preload("Tickets.TicketType").
		Preload("TicketRelease.Event.FormFields").
		Preload("TicketRelease.AddOns").
		Preload("TicketRelease.PaymentDeadline").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("Ticket.EventFormReponses").
		Preload("Ticket.TicketAddOns.AddOn").
		Where("user_ug_kth_id = ?", userUGKthID).
		Find(&ticketOrder)

	if ids != nil {
		if len(*ids) > 0 {
			query = query.Where("id IN (?)", *ids)
		}
	}

	if err := query.Find(&ticketOrder).Error; err != nil {
		return nil, err
	}

	return ticketOrder, nil
}

func GetValidTicketOrder(db *gorm.DB, ticketOrderID uint) (*TicketOrder, error) {
	var ticketOrder TicketOrder
	if err := db.
		Preload("Tickets.TicketType").
		Preload("TicketRelease.Event.FormFields").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("Tickets.EventFormReponses").
		Where("id = ?", ticketOrderID).First(&ticketOrder).Error; err != nil {
		return nil, err
	}

	return &ticketOrder, nil
}

func GetAllValidTicketOrdersToTicketReleaseOrderedByCreatedAt(db *gorm.DB, ticketReleaseID uint) ([]TicketOrder, error) {
	var ticketOrders []TicketOrder
	if err := db.
		Preload("Tickets.TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.PaymentDeadline").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("Tickets.TicketAddOns.AddOn").
		Where("ticket_release_id = ? AND is_handled = ?", ticketReleaseID, false).Order("created_at").Find(&ticketOrders).Error; err != nil {
		return nil, err
	}

	// According to gorm soft delete, we should not fetch soft deleted records

	return ticketOrders, nil
}

func GetAllValidTicketOrdersToTicketRelease(db *gorm.DB, ticketReleaseID uint) ([]TicketOrder, error) {
	var ticketOrders []TicketOrder
	if err := db.
		Preload("Tickets.TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("TicketRelease.PaymentDeadline").
		Preload("Tickets.TicketAddOns.AddOn").
		Where("ticket_release_id = ? AND is_handled = ?", ticketReleaseID, false).Find(&ticketOrders).Error; err != nil {
		return nil, err
	}

	// According to gorm soft delete, we should not fetch soft deleted records

	return ticketOrders, nil
}

func GetTicketOrdersToEvent(db *gorm.DB, eventID uint) ([]TicketOrder, error) {
	var ticketOrders []TicketOrder
	if err := db.
		Preload("Tickets.TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("TicketRelease.PaymentDeadline").
		Preload("Tickets.TicketAddOns.AddOn").
		Joins("JOIN ticket_releases ON ticket_orders.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ?", eventID).Find(&ticketOrders).Error; err != nil {
		return nil, err
	}

	return ticketOrders, nil
}
