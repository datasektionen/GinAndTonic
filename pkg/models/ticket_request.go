package models

import (
	"gorm.io/gorm"
)

type TicketRequest struct {
	gorm.Model
	TicketAmount    int           `json:"ticket_amount"`
	TicketReleaseID uint          `json:"ticket_release_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	TicketRelease   TicketRelease `json:"ticket_release"`
	TicketTypeID    uint          `json:"ticket_type_id" gorm:"index" `
	TicketType      TicketType    `json:"ticket_type"`
	UserUGKthID     string        `json:"user_ug_kth_id"`
	IsHandled       bool          `json:"is_handled" gorm:"default:false"`
}
