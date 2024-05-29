package models

import (
	"time"

	"gorm.io/gorm"
)

type OrganizationStore struct {
	StoreID        string          `json:"store_id" gorm:"primaryKey"`
	Name           string          `json:"name"`
	OrganizationID uint            `json:"organization_id"`
	Terminals      []StoreTerminal `gorm:"foreignKey:StoreID" json:"terminals"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
