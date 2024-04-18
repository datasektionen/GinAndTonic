package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTeamRole(name string) *models.TeamRole {
	return &models.TeamRole{
		Name: name,
	}
}
