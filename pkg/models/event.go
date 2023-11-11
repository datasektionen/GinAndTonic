package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Date           time.Time       `json:"date"`
	Location       string          `json:"location"`
	OrganizationID int             `gorm:"index" json:"organization_id"`
	Organization   Organization    `json:"organization"`
	TicketReleases []TicketRelease `gorm:"foreignKey:EventID" json:"ticket_releases"`
	CreatedBy      string          `json:"created_by"`
}

func GetEvent(db *gorm.DB, id uint) (event Event, err error) {
	err = db.Preload("Organization").First(&event, id).Error
	return
}
