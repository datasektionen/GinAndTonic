package factory

import (
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewOrganizationUserRole(userUGKthID string, organizationID uint, organizationRoleName string) *models.OrganizationUserRole {
	return &models.OrganizationUserRole{
		UserUGKthID:          userUGKthID,
		OrganizationID:       organizationID,
		OrganizationRoleName: organizationRoleName,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}
