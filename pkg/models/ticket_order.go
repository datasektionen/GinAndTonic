package models

import (
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
	Tickets         []Ticket        `gorm:"foreignKey:TicketOrderID" json:"requests"`
	TotalAmount     float64         `json:"total_amount"`
	IsPaid          bool            `json:"is_paid" gorm:"default:false"`
	PaidAt          *time.Time      `json:"paid_at" gorm:"default:null"`
	Type            TicketOrderType `json:"type" gorm:"default:'request'"`
}

func (to *TicketOrder) IsTicketRequest() bool {
	return to.Type == TicketOrderRequest
}

func GetAllValidTicketOrdersToTicketRelease(db *gorm.DB, ticketReleaseID uint) ([]TicketOrder, error) {
	var ticketOrders []TicketOrder
	if err := db.
		Preload("Ticket.TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("TicketRelease.PaymentDeadline").
		Preload("Ticket.TicketAddOns.AddOn").
		Where("ticket_release_id = ? AND is_handled = ?", ticketReleaseID, false).Find(&ticketOrders).Error; err != nil {
		return nil, err
	}

	return ticketOrders, nil
}
