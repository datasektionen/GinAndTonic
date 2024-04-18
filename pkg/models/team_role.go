package models

import (
	"fmt"

	"gorm.io/gorm"
)

type OrgRole string

const (
	TeamMember OrgRole = "member"
	TeamOwner  OrgRole = "owner"
)

func StringToOrgRole(s string) (OrgRole, error) {
	switch s {
	case string(TeamMember):
		return TeamMember, nil
	case string(TeamOwner):
		return TeamOwner, nil
	default:
		return "", fmt.Errorf("invalid team role")
	}
}

type TeamRole struct {
	gorm.Model
	Name string `gorm:"unique" json:"name"`
}

func (TeamRole) TableName() string {
	return "team_roles"
}

func InitializeTeamRoles(db *gorm.DB) error {
	// Define the roles you want to ensure exist
	orgRoles := []TeamRole{
		{Name: string(TeamMember)},
		{Name: string(TeamOwner)},
	}

	// Check each role and create it if it doesn't exist
	for _, orgRole := range orgRoles {
		var existingRole TeamRole
		db.Where("name = ?", orgRole.Name).FirstOrInit(&existingRole)
		if existingRole.ID == 0 {
			if err := db.Create(&orgRole).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func GetTeamRole(db *gorm.DB, role OrgRole) (TeamRole, error) {
	var teamRole TeamRole
	err := db.Where("name = ?", string(role)).First(&teamRole).Error
	return teamRole, err
}
