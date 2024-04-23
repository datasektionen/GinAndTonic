package factory

import (
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewUser(ugKthID, username, firstName, lastName, email string, roleID uint) *models.User {
	return &models.User{
		UGKthID:   ugKthID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		RoleID:    roleID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
