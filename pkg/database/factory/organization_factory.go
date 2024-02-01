package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewOrganization(name, email string) *models.Organization {
	return &models.Organization{
		Name:  name,
		Email: email,
	}
}
