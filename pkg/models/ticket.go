package models

import (
	"gorm.io/gorm"
)

type Ticket struct {
	gorm.Model
	TicketRequestID int           `gorm:"index"`
	TicketRequest   TicketRequest `json:"ticket_request"`
	TicketReleaseID int           `json:"ticket_release_id"`
	TicketRelease   TicketRelease `json:"ticket_release"`
	IsPaid          bool          `json:"is_paid"`
	IsReserve       bool          `json:"is_reserve"`
	UserUGKthID     string        `json:"user_ug_kth_id"`
	User            User          `json:"user"`
}
