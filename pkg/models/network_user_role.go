package models

import "time"

type NetworkUserRole struct {
	UserUGKthID     string `gorm:"primaryKey" json:"ug_kth_id"`
	NetworkID       uint   `gorm:"primaryKey" json:"network_id"`
	NetworkRoleName string `gorm:"primaryKey" json:"network_role_name"`

	CreatedAt time.Time `json:"created_at" default:"CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" default:"CURRENT_TIMESTAMP"`
}
