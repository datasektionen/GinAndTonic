package factory

import (
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewEvent(name, description, location, createdBy string, teamID int, eventDate time.Time) *models.Event {
	return &models.Event{
		Name:        name,
		Description: description,
		Date:        eventDate,
		Location:    location,
		TeamID:      teamID,
		CreatedBy:   createdBy,
	}
}
