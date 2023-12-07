package models

import (
	"time"
)

type OrganizationUserRole struct {
	UserUGKthID          string `gorm:"primaryKey" json:"ug_kth_id"`
	OrganizationID       uint   `gorm:"primaryKey" json:"organization_id"`
	OrganizationRoleName string `gorm:"primaryKey" json:"organization_role_name"`

	CreatedAt time.Time `json:"created_at" default:"CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" default:"CURRENT_TIMESTAMP"`
}
