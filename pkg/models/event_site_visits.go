package models

import "gorm.io/gorm"

type EventSiteVisit struct {
	gorm.Model
	UserUGKthID string `json:"user_ugkth_id"`
	UserAgent   string `json:"user_agent"`
	ReferrerURL string `json:"referrer_url"`
	Location    string `json:"location"`
	EventID     uint   `json:"event_id"` // foreign key field
}
