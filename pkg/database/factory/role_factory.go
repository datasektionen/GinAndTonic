package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewRole(name models.RoleType) *models.Role {
	return &models.Role{
		Name: name,
	}
}
