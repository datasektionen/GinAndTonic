package models

import (
	"errors"

	"gorm.io/gorm"
)

// Notification types

type NotificationType string
type NotificationStatus string

const (
	EmailNotification NotificationType = "email"
)

// Statuses
const (
	PendingNotification NotificationStatus = "pending"
	SentNotification    NotificationStatus = "sent"
	FailedNotification  NotificationStatus = "failed"
)

type Notification struct {
	gorm.Model
	UserUGKthID   string             `json:"user_ug_kth_id"`
	EventID       *uint              `json:"event_id"`
	User          User               `json:"user"`
	Type          NotificationType   `json:"type"`
	Status        NotificationStatus `json:"status" gorm:"default:'pending'"`
	SendOutID     *uint              `json:"send_out_id"`
	Subject       *string            `json:"subject"`
	Content       *string            `json:"content"`
	StatusMessage *string            `json:"status_message"`
}

func (n *Notification) Validate() error {
	if n.Type != EmailNotification {
		return errors.New("invalid notification type")
	}
	return nil
}

// Func that sets the status of a notification to failed and then saves it to the database
func (n *Notification) SetFailed(db *gorm.DB, err error) error {
	n.Status = FailedNotification

	errorMessage := err.Error()
	n.StatusMessage = &errorMessage
	return db.Save(n).Error
}

func (n *Notification) SetSent(db *gorm.DB) error {
	n.Status = SentNotification
	return db.Save(n).Error
}
