package factory

import (
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTeamUserRole(userUGKthID string, teamID uint, teamRoleName string) *models.TeamUserRole {
	return &models.TeamUserRole{
		UserUGKthID:  userUGKthID,
		TeamID:       teamID,
		TeamRoleName: teamRoleName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
