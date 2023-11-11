package models

import (
	"gorm.io/gorm"
)

type Ticket struct {
	gorm.Model
	TicketRequestID uint          `gorm:"index"`
	TicketRequest   TicketRequest `json:"ticket_request"`
	IsPaid          bool          `json:"is_paid"`
	IsReserve       bool          `json:"is_reserve"`
	UserUGKthID     string        `json:"user_ug_kth_id"`
	User            User          `json:"user"`
}
