package models

import (
	"errors"
	"time"

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
	TicketsAvailable            int                       `json:"tickets_available"`
	IsReserved                  bool                      `json:"is_reserved" default:"false"`
	PromoCode                   *string                   `gorm:"default:NULL" json:"promo_code"`
	PayWithin                   *int64                    `json:"pay_within" default:"NULL"`
	HasAllocatedTickets         bool                      `json:"has_allocated_tickets"`
	TicketReleaseMethodDetailID uint                      `gorm:"index" json:"ticket_release_method_detail_id"`
	TicketReleaseMethodDetail   TicketReleaseMethodDetail `json:"ticket_release_method_detail"`
	ReservedUsers               []User                    `gorm:"many2many:user_unlocked_ticket_releases;" json:"-"`
}

func (tr *TicketRelease) ValidatePayWithin() bool {
	payWithin := *tr.PayWithin
	println(payWithin)

	if tr.PayWithin == nil {
		println("PayWithin is not nil")
		return false
	}

	if tr.PayWithin != nil && *tr.PayWithin < 0 {
		println("PayWithin is not less than 0")
		return false
	}

	println("Event date: ", tr.Event.Date.Unix())

	if tr.Event.Date.Unix() < time.Now().Add(time.Duration(*tr.PayWithin)*time.Hour).Unix() {
		println("Event date is not less than pay within")
		return false
	}

	return true
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

func (tr *TicketRelease) UserHasAccessToTicketRelease(DB *gorm.DB, id string) bool {
	if !tr.IsReserved {
		return true
	}

	var user User
	if err := DB.
		Preload("Organizations").Where("ug_kth_id = ?", id).First(&user).Error; err != nil {
		return false
	}

	for _, reservedUser := range tr.ReservedUsers {
		if reservedUser.UGKthID == user.UGKthID {
			return true
		}
	}

	if tr.Event.OrganizationID == 0 {
		panic("Need to prelaod organization")
	}

	// Check if user is in same organization as ticket release
	for _, organization := range user.Organizations {
		if organization.ID == uint(tr.Event.OrganizationID) {
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

func GetOpenTicketReleases(db *gorm.DB) (ticketReleases []TicketRelease, err error) {
	err = db.Where("open <= ? AND close >= ?", time.Now().Unix(), time.Now().Unix()).Find(&ticketReleases).Error

	if err != nil {
		return nil, err
	}

	return ticketReleases, nil
}

func GetClosedTicketReleases(db *gorm.DB) (ticketReleases []TicketRelease, err error) {
	err = db.Where("close <= ?", time.Now().Unix()).Find(&ticketReleases).Error

	if err != nil {
		return nil, err
	}

	return ticketReleases, nil
}
