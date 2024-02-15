package models

import (
	"time"

	"gorm.io/gorm"
)

// Event is a struct that represents an event in the database
type Event struct {
	gorm.Model
	Name           string          `json:"name"`
	Description    string          `json:"description" gorm:"type:text"`
	Date           time.Time       `json:"date"`
	Location       string          `json:"location"`
	OrganizationID int             `gorm:"index" json:"organization_id"`
	Organization   Organization    `json:"organization"`
	TicketReleases []TicketRelease `gorm:"foreignKey:EventID" json:"ticket_releases"`
	IsPrivate      bool            `json:"is_private" gorm:"default:false"`
	SecretToken    string          `json:"-"`
	CreatedBy      string          `json:"created_by"`
}

// GetEvent returns an event from the database
func GetEvent(db *gorm.DB, id uint) (event Event, err error) {
	err = db.Preload("Organization").First(&event, id).Error
	return
}
