package models

import (
	"gorm.io/gorm"
)

type TicketType struct {
	gorm.Model
	EventID         uint    `gorm:"index" json:"event_id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Price           float64 `json:"price"`
	TicketReleaseID uint    `json:"ticket_release_id"`
}

func (tt *TicketType) IsFree() bool {
	return tt.Price == 0
}
