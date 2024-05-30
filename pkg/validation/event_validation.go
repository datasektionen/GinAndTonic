package validation

import (
	"errors"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func ValidateEventDates(db *gorm.DB, eventID uint) error {
	var event models.Event
	if err := db.Preload("TicketReleases").First(&event, eventID).Error; err != nil {
		return err
	}

	ticketReleases := event.TicketReleases

	for _, ticketRelease := range ticketReleases {
		// Check if event.date is after ticketRelease.close
		// Convert from unix to time.Time

		if event.Date.Before(ticketRelease.Close) {
			return errors.New("event date is after ticket release close")
		}

		if event.EndDate != nil {
			if event.EndDate.Before(ticketRelease.Close) {
				return errors.New("event end date is after ticket release close")
			}
		}

		if err := ticketRelease.ValidateTicketReleaseDates(db); err != nil {
			return err
		}
	}

	return nil
}
