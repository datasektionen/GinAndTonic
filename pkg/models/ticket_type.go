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
	SaveTemplate    bool    `json:"save_template" gorm:"default:false"`
}

// IsFree returns true if the ticket type is free
func (tt *TicketType) IsFree() bool {
	return tt.Price == 0
}

func (tt *TicketType) UserHasAccessToTicketType(db *gorm.DB, user User) bool {
	var ticketRelease TicketRelease
	db.First(&ticketRelease, tt.TicketReleaseID)

	var event Event
	db.First(&event, ticketRelease.EventID)

	var organization Organization
	db.First(&organization, event.OrganizationID)

	for _, userOrganization := range user.Organizations {
		if userOrganization.ID == organization.ID {
			return true
		}
	}

	return false
}
