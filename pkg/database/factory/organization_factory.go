package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTeam(name, email string) *models.Team {
	return &models.Team{
		Name:  name,
		Email: email,
	}
}
