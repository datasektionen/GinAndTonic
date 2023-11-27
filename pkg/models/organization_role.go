package models

import (
	"gorm.io/gorm"
)

type OrgRole string

const (
	OrganizationMember OrgRole = "member"
	OrganizationOwner  OrgRole = "owner"
)

type OrganizationRole struct {
	gorm.Model
	Name string `gorm:"unique" json:"name"`
}

func InitializeOrganizationRoles(db *gorm.DB) error {
	// Define the roles you want to ensure exist
	orgRoles := []OrganizationRole{
		{Name: string(OrganizationMember)},
		{Name: string(OrganizationOwner)},
	}

	// Check each role and create it if it doesn't exist
	for _, orgRole := range orgRoles {
		var existingRole OrganizationRole
		db.Where("name = ?", orgRole.Name).FirstOrInit(&existingRole)
		if existingRole.ID == 0 {
			if err := db.Create(&orgRole).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func GetOrganizationRole(db *gorm.DB, role OrgRole) (OrganizationRole, error) {
	var organizationRole OrganizationRole
	err := db.Where("name = ?", string(role)).First(&organizationRole).Error
	return organizationRole, err
}
