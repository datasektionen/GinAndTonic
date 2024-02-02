package models

import (
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name string `gorm:"unique" json:"name"`

	Users []User `gorm:"foreignKey:RoleID" json:"users"`
}

func GetRole(db *gorm.DB, name string) (Role, error) {
	var role Role
	err := db.Where("name = ?", name).First(&role).Error
	return role, err
}

func InitializeRoles(db *gorm.DB) error {
	// Define the roles you want to ensure exist
	roles := []Role{
		{Name: "super_admin"},
		{Name: "user"},
		{Name: "external"},
	}

	// Check each role and create it if it doesn't exist
	for _, role := range roles {
		var existingRole Role
		db.Where("name = ?", role.Name).FirstOrInit(&existingRole)
		if existingRole.ID == 0 {
			if err := db.Create(&role).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
