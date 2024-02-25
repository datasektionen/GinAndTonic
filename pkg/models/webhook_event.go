package models

import (
	"gorm.io/gorm"
)

type WebhookEvent struct {
	gorm.Model
	StripeID  string `gorm:"uniqueIndex;type:varchar(255)"` // Stripe Event ID
	EventType string `json:"event_type" gorm:"index"`       // Type of event
	LastError string `json:"last_error" gorm:"type:text"`
	Processed bool   `json:"processed" gorm:"default:false"` // If the event has been processed
}
