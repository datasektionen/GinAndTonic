package models

import "gorm.io/gorm"

type ReferralSource struct {
	gorm.Model
	UserUGKthID string `json:"user_ug_kth_id"`
	Source      string `json:"source"`
	Specific    string `json:"specific"`
}
