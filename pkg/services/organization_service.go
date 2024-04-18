package services

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type OrganisationService struct {
	DB *gorm.DB
}

func NewTeamService(db *gorm.DB) *OrganisationService {
	return &OrganisationService{DB: db}
}

func (os *OrganisationService) AddUserToTeam(username string, teamID uint, teamRole models.OrgRole) error {
	var user models.User
	var team models.Team

	if err := os.DB.First(&user, "username = ?", username).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	if err := os.DB.First(&team, teamID).Error; err != nil {
		return fmt.Errorf("team not found")
	}

	// Start transaction
	tx := os.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Associate user with team
	if err := tx.Model(&team).Association("Users").Append(&user); err != nil {
		tx.Rollback()
		return err
	}
	// 2. Create team user role
	teamUserRole := models.TeamUserRole{
		UserUGKthID:  user.UGKthID,
		TeamID:       team.ID,
		TeamRoleName: string(teamRole),
	}

	if err := tx.Create(&teamUserRole).Error; err != nil {
		tx.Rollback()
		return err
	}
	// Commit transaction
	tx.Commit()

	return nil
}

func (os *OrganisationService) RemoveUserFromTeam(username string, teamID uint) error {
	var user models.User
	var team models.Team

	if err := os.DB.First(&user, "username = ?", username).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	if err := os.DB.Preload("Users").Preload("Users.TeamUserRoles").First(&team, teamID).Error; err != nil {
		return fmt.Errorf("team not found")
	}

	isOwner, err := os.isUserOwnerOfTeam(user.UGKthID, team)
	if err != nil {
		return err
	}

	if isOwner && len(team.Users) == 1 {
		return fmt.Errorf("user %v is the owner of the team %v", username, teamID)
	}

	if err := os.DB.Model(&team).Association("Users").Delete(&user); err != nil {
		return fmt.Errorf("there was an error removing the user from the team: %w", err)
	}

	// Remove the user.TeamUserRole for this team
	if err := os.DB.Unscoped().Where("user_ug_kth_id = ? AND team_id = ?", user.UGKthID, team.ID).Delete(&models.TeamUserRole{}).Error; err != nil {
		return fmt.Errorf("there was an error removing the user from the team: %w", err)
	}

	return nil
}

func (os *OrganisationService) GetTeamUsers(teamID uint) ([]models.User, error) {
	var team models.Team

	if err := os.DB.First(&team, teamID).Error; err != nil {
		return nil, fmt.Errorf("team not found")
	}

	users, err := team.GetUsers(os.DB)

	if err != nil {
		return nil, fmt.Errorf("there was an error fetching the team users: %w", err)
	}

	return users, nil
}

func (os *OrganisationService) isUserOwnerOfTeam(userUGKthID string, team models.Team) (bool, error) {
	orgOwners, err := models.GetTeamOwners(os.DB, team)
	if err != nil {
		return false, fmt.Errorf("there was an error fetching the team owners: %w", err)
	}

	for _, owner := range orgOwners {
		if owner.UGKthID == userUGKthID {
			return true, nil
		}
	}

	return false, nil
}

func (os *OrganisationService) ChangeUserRoleInTeam(username string, teamID uint, newRole models.OrgRole) error {
	var user models.User
	var team models.Team

	// Find the user
	if err := os.DB.First(&user, "username = ?", username).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	// Find the team
	if err := os.DB.First(&team, teamID).Error; err != nil {
		return fmt.Errorf("team not found")
	}

	// Find the team user role
	var teamUserRole models.TeamUserRole
	if err := os.DB.Where("user_ug_kth_id = ? AND team_id = ?", user.UGKthID, team.ID).First(&teamUserRole).Error; err != nil {
		return fmt.Errorf("team user role not found")
	}

	// Start transaction
	tx := os.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the role
	if err := tx.Model(&teamUserRole).Update("team_role_name", string(newRole)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("there was an error updating the user role: %w", err)
	}

	// Commit transaction
	tx.Commit()

	return nil
}
