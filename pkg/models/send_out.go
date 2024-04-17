package models

import "gorm.io/gorm"

type SendOut struct {
	gorm.Model
	EventID       *uint          `json:"event_id" gorm:"default:NULL"`
	Notifications []Notification `json:"notifications" gorm:"foreignKey:SendOutID"`
	Subject       string         `json:"subject"`
	Content       string         `json:"content"`
}
