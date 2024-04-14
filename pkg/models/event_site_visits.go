package models

import (
	"gorm.io/gorm"
)

type EventSiteVisit struct {
	gorm.Model
	UserUGKthID string `json:"user_ugkth_id"`
	UserAgent   string `json:"user_agent"`
	ReferrerURL string `json:"referrer_url"`
	Location    string `json:"-"`
	EventID     uint   `json:"event_id"` // foreign key field
}

type EventSiteVisitSummary struct {
	gorm.Model
	EventID           uint    `json:"event_id"`
	TotalVisits       int     `json:"total_visits"`
	UniqueUsers       int     `json:"unique_users"`
	NumTicketRequests int     `json:"num_ticket_requests"`
	TotalIncome       float64 `json:"total_income"`
}

// Get EventSiteVisits of all events that has passed
