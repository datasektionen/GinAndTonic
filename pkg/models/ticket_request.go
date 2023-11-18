package models

import (
	"gorm.io/gorm"
)

type TicketRequest struct {
	gorm.Model
	TicketAmount    int           `json:"ticket_amount"`
	TicketReleaseID uint          `gorm:"index" json:"ticket_release_id"`
	TicketRelease   TicketRelease `json:"ticket_release"`
	TicketTypeID    uint          `gorm:"index" json:"ticket_type_id"`
	TicketType      TicketType    `json:"ticket_type"`
	UserUGKthID     string        `json:"user_ug_kth_id"`
	IsHandled       bool          `json:"is_handled" gorm:"default:false"`
}
