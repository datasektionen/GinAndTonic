package models

import (
	"time"

	"gorm.io/gorm"
)

// Event is a struct that represents an event in the database
type Event struct {
	gorm.Model
	Name                 string           `json:"name"`
	Description          string           `json:"description" gorm:"type:text"`
	Date                 time.Time        `json:"date"`
	Location             string           `json:"location"`
	OrganizationID       int              `gorm:"index" json:"organization_id"`
	Organization         Organization     `json:"organization"`
	TicketReleases       []TicketRelease  `gorm:"foreignKey:EventID" json:"ticket_releases"`
	IsPrivate            bool             `json:"is_private"`
	SecretToken          string           `json:"-"`
	CreatedBy            string           `json:"created_by"`
	FormFieldDescription *string          `json:"form_field_description" gorm:"type:text"`
	FormFields           []EventFormField `gorm:"foreignKey:EventID" json:"form_fields"`
}

// GetEvent returns an event from the database
func GetEvent(db *gorm.DB, id uint) (event Event, err error) {
	err = db.Preload("Organization").First(&event, id).Error
	return
}

// Func get all ticket releases to event
func GetTicketReleasesToEvent(db *gorm.DB, eventID uint) (ticketReleases []TicketRelease, err error) {
	err = db.Where("event_id = ?", eventID).Find(&ticketReleases).Error
	return
}
