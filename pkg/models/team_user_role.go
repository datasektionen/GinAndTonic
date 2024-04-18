package models

import (
	"time"
)

type TeamUserRole struct {
	UserUGKthID  string `gorm:"primaryKey" json:"ug_kth_id"`
	TeamID       uint   `gorm:"primaryKey" json:"team_id"`
	TeamRoleName string `gorm:"primaryKey" json:"team_role_name"`

	CreatedAt time.Time `json:"created_at" default:"CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" default:"CURRENT_TIMESTAMP"`
}

func (TeamUserRole) TableName() string {
	return "team_user_roles"
}
