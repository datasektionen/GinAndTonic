package models

import (
	"gorm.io/gorm"
)

type EventSiteVisit struct {
	gorm.Model
	EventID     uint   `json:"event_id" gorm:"index"`
	UserAgent   string `json:"user_agent"`
	ReferrerURL string `json:"referrer_url"`
	Location    string `json:"-"`
}

type EventSiteVisitSummary struct {
	gorm.Model
	EventID     uint    `json:"event_id" gorm:"index"`
	TotalVisits int     `json:"total_visits"`
	UniqueUsers int     `json:"unique_users"`
	NumTickets  int     `json:"num_tickets"`
	TotalIncome float64 `json:"total_income"`
}

// Get EventSiteVisits of all events that has passed
