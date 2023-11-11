package models

import (
	"time"

	"gorm.io/gorm"
)

type OrganizationUserRole struct {
	UserUGKthID          string `gorm:"primaryKey"`
	OrganizationID       uint   `gorm:"primaryKey"`
	OrganizationRoleName string `gorm:"primaryKey"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
