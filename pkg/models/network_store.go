package models

import (
	"time"

	"gorm.io/gorm"
)

type NetworkStore struct {
	StoreID        string            `json:"store_id" gorm:"primaryKey"`
	OrganizationID uint              `json:"organization_id"`
	Terminals      []NetworkTerminal `gorm:"foreignKey:StoreID" json:"terminals"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
