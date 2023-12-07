package models

import (
	"errors"
	"regexp"

	"gorm.io/gorm"
)

type Organization struct {
	gorm.Model
	Name                  string                 `gorm:"uniqueIndex" json:"name"`
	Events                []Event                `gorm:"foreignKey:OrganizationID" json:"events"`
	Users                 []User                 `gorm:"many2many:organization_users;" json:"users"`
	OrganizationUserRoles []OrganizationUserRole `gorm:"foreignKey:OrganizationID" json:"organization_user_roles"`
}

func (o Organization) ValidateName() error {
	if len(o.Name) < 3 {
		return errors.New("name must be at least 3 characters long")
	}
	if len(o.Name) > 50 {
		return errors.New("name must not be longer than 50 characters")
	}
	if match, _ := regexp.MatchString(`^[a-zA-Z0-9]*$`, o.Name); !match {
		return errors.New("name must not contain special characters or spaces")
	}
	return nil
}

func (o Organization) Validate() error {
	if err := o.ValidateName(); err != nil {
		return err
	}
	return nil
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

func (o Organization) GetUsers(db *gorm.DB) ([]User, error) {
	var users []User

	// Preload the Users and OrganizationUserRoles for the organization
	if err := db.Model(&o).Preload("OrganizationUserRoles").Association("Users").Find(&users); err != nil {
		return nil, err
	}

	// Remove the user.OrganizationUserRoles that are not for this organization
	for i, user := range users {
		var roles []OrganizationUserRole
		for _, role := range user.OrganizationUserRoles {
			if role.OrganizationID == o.ID {
				roles = append(roles, role)
			}
		}
		users[i].OrganizationUserRoles = roles
	}

	return users, nil
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
