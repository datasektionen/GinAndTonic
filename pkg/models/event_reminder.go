package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TicketReleaseReminder struct {
	gorm.Model
	TicketReleaseID uint          `json:"ticket_release_id" gorm:"index"`
	TicketRelease   TicketRelease `json:"ticket_release"`
	UserUGKthID     string        `json:"user_ug_kth_id" gorm:"index"`
	User            User          `json:"user"`
	ReminderTime    time.Time     `json:"reminder_time" gorm:"index"`
	IsSent          bool          `gorm:"default:false"`
}

func (trr *TicketReleaseReminder) Validate(db *gorm.DB) error {
	// Check if the pair of ticket release and user exists
	var existingReminder TicketReleaseReminder
	if err := db.Where("ticket_release_id = ? AND user_ug_kth_id = ?", trr.TicketReleaseID, trr.UserUGKthID).First(&existingReminder).Error; err == nil {
		return fmt.Errorf("reminder already exists")
	}

	var ticketRelease TicketRelease
	if err := db.First(&ticketRelease, trr.TicketReleaseID).Error; err != nil {
		return fmt.Errorf("ticket release not found")
	}

	if time.Now().After(trr.ReminderTime) {
		return fmt.Errorf("reminder time is in the past")
	}

	if trr.ReminderTime.After(ticketRelease.Open) {
		return fmt.Errorf("reminder time is after the ticket release opens")
	}

	return nil
}

func CreateTicketReleaseReminder(db *gorm.DB, trReminder *TicketReleaseReminder) error {
	if err := db.Create(trReminder).Error; err != nil {
		return err
	}

	return nil
}
