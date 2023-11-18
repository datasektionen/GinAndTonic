package factory

import (
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewEvent(name, description, location, createdBy string, organizationID int, eventDate time.Time) *models.Event {
	return &models.Event{
		Name:           name,
		Description:    description,
		Date:           eventDate,
		Location:       location,
		OrganizationID: organizationID,
		CreatedBy:      createdBy,
	}
}
