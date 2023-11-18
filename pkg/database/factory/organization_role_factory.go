package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewOrganizationRole(name string) *models.OrganizationRole {
	return &models.OrganizationRole{
		Name: name,
	}
}
