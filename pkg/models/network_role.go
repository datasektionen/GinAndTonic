package models

import (
	"gorm.io/gorm"
)

type NetRole string

const (
	NetworkSuperAdmin NetRole = "network_super_admin"
	NetworkAdmin      NetRole = "network_admin"
	NetworkMember     NetRole = "network_member"
)

type NetworkRole struct {
	gorm.Model
	Name string `gorm:"unique" json:"name"`
}

func InitializeNetworkRoles(db *gorm.DB) error {
	// Define the roles you want to ensure exist
	netRoles := []NetworkRole{
		{Name: string(NetworkSuperAdmin)},
		{Name: string(NetworkAdmin)},
		{Name: string(NetworkMember)},
	}

	// Check each role and create it if it doesn't exist
	for _, netRole := range netRoles {
		var existingRole NetworkRole
		db.Where("name = ?", netRole.Name).FirstOrCreate(&existingRole, netRole)
	}
	return nil
}

func GetNetworkRoleByName(db *gorm.DB, name string) (NetworkRole, error) {
	var role NetworkRole
	err := db.Where("name = ?", name).First(&role).Error
	return role, err
}
