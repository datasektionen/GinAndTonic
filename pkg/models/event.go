package models

import (
	"time"

	"gorm.io/gorm"
)

// Event is a struct that represents an event in the database
type Event struct {
	gorm.Model
	Name                 string                  `json:"name"`
	Description          string                  `json:"description" gorm:"type:text"`
	Date                 time.Time               `json:"date"`
	EndDate              *time.Time              `json:"end_date" gorm:"default:null"`
	Location             string                  `json:"location"`
	OrganizationID       int                     `gorm:"index" json:"organization_id"`
	Organization         Organization            `json:"organization"`
	TicketReleases       []TicketRelease         `gorm:"foreignKey:EventID" json:"ticket_releases"`
	IsPrivate            bool                    `json:"is_private"`
	SecretToken          string                  `json:"-"`
	CreatedBy            string                  `json:"created_by"`
	FormFieldDescription *string                 `json:"form_field_description" gorm:"type:text"`
	FormFields           []EventFormField        `gorm:"foreignKey:EventID" json:"form_fields"`
	SiteVisits           []EventSiteVisit        `gorm:"foreignKey:EventID" json:"-"`
	SiteVisitSummaries   []EventSiteVisitSummary `gorm:"foreignKey:EventID" json:"-"`
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

// Get Events that are in the future
func GetFutureEvents(db *gorm.DB) (events []Event, err error) {
	now := time.Now()
	err = db.Find(&events).Error

	if err != nil {
		return nil, err
	}

	for _, event := range events {
		if event.EndDate != nil {
			if event.EndDate.Before(now) {
				events = append(events, event)
			}
		}

		if event.Date.Before(now) {
			events = append(events, event)
		}
	}

	return events, nil
}

func GetEventSiteVisits(db *gorm.DB, eventID uint) (eventSiteVisits []EventSiteVisit, err error) {
	// Check that Event.Date is in the past or if Event.EndDate is in the past
	var event Event
	err = db.Preload("SiteVisits").Find(&event, eventID).Error
	if err != nil {
		return nil, err
	}

	eventSiteVisits = append(eventSiteVisits, event.SiteVisits...)

	return eventSiteVisits, nil
}
