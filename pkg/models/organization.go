package models

import (
	"gorm.io/gorm"
)

type Organization struct {
	gorm.Model
	Name                  string                 `gorm:"uniqueIndex" json:"name"`
	Events                []Event                `gorm:"foreignKey:OrganizationID" json:"events"`
	Users                 []User                 `gorm:"many2many:organization_users;" json:"users"`
	OrganizationUserRoles []OrganizationUserRole `gorm:"foreignKey:OrganizationID" json:"organization_user_roles"`
}

func GetOrganizationByNameIfExist(db *gorm.DB, name string) (Organization, error) {
	var organization Organization
	err := db.Preload("Events").Where("name = ?", name).First(&organization).Error
	return organization, err
}

func GetOrganizationByIDIfExist(db *gorm.DB, id uint) (Organization, error) {
	var organization Organization
	err := db.Preload("Events").Where("id = ?", id).First(&organization).Error
	return organization, err
}

func GetOrganizationOwners(db *gorm.DB, organization Organization) (users []User, err error) {
	for _, user := range organization.Users {
		for _, role := range user.OrganizationUserRoles {

			if role.OrganizationRoleName == string(OrganizationOwner) && role.OrganizationID == organization.ID {
				users = append(users, user)
			}
		}
	}

	return users, nil
}
