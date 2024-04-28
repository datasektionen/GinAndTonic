package models

import (
	"gorm.io/gorm"
)

type NetRole string

const (
	NetworkNode NetRole = "node"
	NetworkRoot NetRole = "root"
)

type NetworkRole struct {
	gorm.Model
	Name string `gorm:"unique" json:"name"`
}

func InitializeNetworkRoles(db *gorm.DB) error {
	// Define the roles you want to ensure exist
	netRoles := []NetworkRole{
		{Name: string(NetworkNode)},
		{Name: string(NetworkRoot)},
	}

	// Check each role and create it if it doesn't exist
	for _, netRole := range netRoles {
		var existingRole NetworkRole
		db.Where("name = ?", netRole.Name).FirstOrCreate(&existingRole, netRole)
	}
	return nil
}
