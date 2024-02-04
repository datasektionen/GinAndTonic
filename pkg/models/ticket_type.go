package models

import (
	"gorm.io/gorm"
)

// TicketType is a struct that represents a ticket type in the database
type TicketType struct {
	gorm.Model
	EventID         uint    `gorm:"index" json:"event_id"`
	Name            string  `json:"name"`
	Description     string  `json:"description" gorm:"type:text"`
	Price           float64 `json:"price"`
	TicketReleaseID uint    `json:"ticket_release_id"`
}

// IsFree returns true if the ticket type is free
func (tt *TicketType) IsFree() bool {
	return tt.Price == 0
}
