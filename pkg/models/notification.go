package models

import (
	"errors"

	"gorm.io/gorm"
)

// Notification types
const (
	EmailNotification = "email"
)

// Statuses
const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
)

type Notification struct {
	gorm.Model
	UserUGKthID string `json:"user_ug_kth_id"`
	User        User   `json:"user"`
	Type        string `json:"type"`
	Subject     string `json:"subject"`
	Status      string `json:"status" gorm:"default:'pending'"`
	Content     string `json:"content"` // Contains the email body in html format
	EventID     *uint  `json:"event_id" gorm:"default:NULL"`
}

func (n *Notification) Validate() error {
	if n.Type != EmailNotification {
		return errors.New("invalid notification type")
	}
	return nil
}
