package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WebhookEvent struct {
	gorm.Model
	EventType        string         `gorm:"index" json:"event_type"`
	EventID          string         `gorm:"index;unique" json:"event_id"`
	WebhookCreatedAt time.Time      `json:"webhook_created_at"`
	RetryAttempt     int            `json:"retry_attempt"`
	WebhookEventID   string         `gorm:"index" json:"webhook_event_id"`
	Payload          datatypes.JSON `json:"payload"` // Generic payload for additional data
}

func (we *WebhookEvent) Process() error {
	// Implement generic processing logic
	return nil
}
