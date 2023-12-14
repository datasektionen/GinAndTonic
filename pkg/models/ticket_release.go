package models

import (
	"errors"

	"gorm.io/gorm"
)

type TicketRelease struct {
	gorm.Model
	EventID                     int                       `gorm:"index" json:"event_id"`
	Event                       Event                     `json:"event"`
	Name                        string                    `json:"name"`
	Description                 string                    `json:"description"`
	Open                        int64                     `json:"open"`
	Close                       int64                     `json:"close"`
	TicketTypes                 []TicketType              `gorm:"foreignKey:TicketReleaseID" json:"ticket_types"`
	TicketRequests              []TicketRequest           `gorm:"foreignKey:TicketReleaseID" json:"ticket_requests"`
	IsReserved                  bool                      `json:"is_reserved" default:"false"`
	PromoCode                   *string                   `gorm:"default:NULL" json:"promo_code"`
	HasAllocatedTickets         bool                      `json:"has_allocated_tickets"`
	TicketReleaseMethodDetailID uint                      `gorm:"index" json:"ticket_release_method_detail_id"`
	TicketReleaseMethodDetail   TicketReleaseMethodDetail `json:"ticket_release_method_detail"`
	ReservedUsers               []User                    `gorm:"many2many:user_unlocked_ticket_releases;" json:"-"`
}

func DeleteTicketRelease(db *gorm.DB, ticketReleaseID uint) error {
	// Begin a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Soft delete associated TicketRequests
	if err := tx.Where("ticket_release_id = ?", ticketReleaseID).Delete(&TicketRequest{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Soft delete TicketRelease
	if err := tx.Delete(&TicketRelease{}, ticketReleaseID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

func (tr *TicketRelease) UserHasAccessToTicketRelease(user *User) bool {
	if !tr.IsReserved {
		return true
	}

	for _, reservedUser := range tr.ReservedUsers {
		if reservedUser.UGKthID == user.UGKthID {
			return true
		}
	}

	return false
}

func (tr *TicketRelease) ValidateTicketReleaseDates(db *gorm.DB) error {
	// Check if ticketRelease.open is before ticketRelease.close
	if tr.Open > tr.Close {
		return errors.New("ticket release open is after ticket release close")
	}

	return nil
}

func (tr *TicketRelease) UserUnlockReservedTicketRelease(user *User) {
	tr.ReservedUsers = append(tr.ReservedUsers, *user)
}
