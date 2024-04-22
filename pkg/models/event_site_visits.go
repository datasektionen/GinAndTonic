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
	EventID           uint    `json:"event_id" gorm:"index"`
	TotalVisits       int     `json:"total_visits"`
	UniqueUsers       int     `json:"unique_users"`
	NumTicketRequests int     `json:"num_ticket_requests"`
	TotalIncome       float64 `json:"total_income"`
}

// Get EventSiteVisits of all events that has passed
