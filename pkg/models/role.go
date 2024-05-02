package models

import (
	"gorm.io/gorm"
)

type RoleType string

const (
	RoleSuperAdmin    RoleType = "super_admin"
	RoleManager       RoleType = "manager"
	RoleCustomer      RoleType = "customer"
	RoleCustomerGuest RoleType = "customer_guest"
)

type Role struct {
	gorm.Model
	Name RoleType `gorm:"unique" json:"name"`

	Users []User `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:RoleID;References:UGKthID;joinReferences:UserUGKthID" json:"users"`
}

func GetRole(db *gorm.DB, name RoleType) (Role, error) {
	var role Role
	err := db.Where("name = ?", name).First(&role).Error
	return role, err
}

func InitializeRoles(db *gorm.DB) error {
	// Define the roles you want to ensure exist
	roles := []Role{
		{Name: RoleSuperAdmin},
		{Name: RoleManager},
		{Name: RoleCustomer},
		{Name: RoleCustomerGuest},
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
