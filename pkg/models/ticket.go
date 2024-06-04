package models

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TicketStatus string

const (
	Pending   TicketStatus = "pending"
	Reserved  TicketStatus = "reserve"
	Cancelled TicketStatus = "cancelled"
	Allocated TicketStatus = "allocated"
)



type Ticket struct {
	gorm.Model

	TicketOrderID     uint                     `json:"ticket_order_id" gorm:"index"`
	TicketOrder       TicketOrder              `json:"ticket_order"`
	TicketTypeID      uint                     `json:"ticket_type_id" gorm:"index"`
	TicketType        TicketType               `json:"ticket_type"`
	UserUGKthID       string                   `json:"user_ug_kth_id"`
	User              User                     `json:"user"`
	TicketAmount      int                      `json:"ticket_amount"`
	IsHandled         bool                     `json:"is_handled" gorm:"default:false"`
	EventFormReponses []EventFormFieldResponse `json:"event_form_responses"`
	TicketAddOns      []TicketAddOn            `gorm:"foreignKey:TicketID" json:"ticket_add_ons"`
	HandledAt         sql.NullTime             `json:"handled_at" gorm:"default:null"`
	DeletedReason     string                   `json:"deleted_reason" gorm:"default:null"`

	// Original Ticket fields
	IsPaid          bool         `json:"is_paid" default:"false"`
	IsReserve       bool         `json:"is_reserve"`
	WasReserve      bool         `json:"was_reserve" default:"false"`
	ReserveNumber   uint         `json:"reserve_number" default:"0"`
	Refunded        bool         `json:"refunded" default:"false"`
	Status          TicketStatus `json:"status" gorm:"default:'pending'"`
	CheckedIn       bool         `json:"checked_in" default:"false"`
	CheckedInAt     sql.NullTime `json:"checked_in_at"`
	QrCode          string       `json:"qr_code" gorm:"unique;not null"`
	PurchasableAt   sql.NullTime `json:"purchasable_at" gorm:"default:null"`
	PaymentDeadline sql.NullTime `json:"payment_deadline" gorm:"default:null"`

	OrderID *string `json:"order_id" gorm:"default:null"`
	Order   Order   `json:"order"`
}

func (t *Ticket) BeforeSave(tx *gorm.DB) (err error) {
	if t.IsHandled && t.HandledAt.Valid {
		now := time.Now()
		t.HandledAt = sql.NullTime{Time: now, Valid: true}
	}
	if t.IsReserve && !t.WasReserve {
		t.WasReserve = true
	}

	return
}
func (t *Ticket) BeforeUpdate(tx *gorm.DB) (err error) {
	if !t.DeletedAt.Valid {
		t.DeletedReason = ""
	}
	return
}

func (t *Ticket) Delete(db *gorm.DB, reason string) error {
	// Delete the associated TicketRequest
	if err := db.Model(t).Update("deleted_reason", reason).Error; err != nil {
		return err
	}

	return db.Delete(t).Error
}

func GetTicketByID(db *gorm.DB, ticketID uint) (ticket Ticket, err error) {
	err = db.
		Preload("TicketType").
		First(&ticket, ticketID).Error
	if err != nil {
		return Ticket{}, err
	}

	return ticket, nil
}

func GetTicketsByIDs(db *gorm.DB, ticketIDs []uint) (tickets []Ticket, err error) {
	err = db.
		Preload("TicketAddOns").
		Preload("TicketType").
		Preload("TicketOrder.TicketRelease").
		Find(&tickets, ticketIDs).Error
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetTicketsToEvent(db *gorm.DB, eventID uint) (tickets []Ticket, err error) {
	// eventID is fetched in TicketOrder.TicketRelease.EventID
	err = db.
		Joins("JOIN ticket_orders ON tickets.ticket_order_id = ticket_orders.id").
		Joins("JOIN ticket_releases ON ticket_orders.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ?", eventID).
		Find(&tickets).Error

	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetAllPaidTicketsToEvent(db *gorm.DB, eventID uint) (tickets []Ticket, err error) {
	// eventID is fetched in TicketOrder.TicketRelease.EventID
	err = db.
		Joins("JOIN ticket_orders ON tickets.ticket_order_id = ticket_orders.id").
		Joins("JOIN ticket_releases ON ticket_orders.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ? AND tickets.is_paid = ?", eventID, true).
		Find(&tickets).Error

	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetTicketToEvent(db *gorm.DB, eventID, ticketID uint) (ticket Ticket, err error) {
	// eventID is fetched in TicketOrder.TicketRelease.EventID
	err = db.
		Joins("JOIN ticket_orders ON tickets.ticket_order_id = ticket_orders.id").
		Joins("JOIN ticket_releases ON ticket_orders.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ? AND tickets.id = ?", eventID, ticketID).
		First(&ticket).Error

	if err != nil {
		return Ticket{}, err
	}

	return ticket, nil

}

func GetAllValidUsersTicket(db *gorm.DB, userUGKthID string) ([]Ticket, error) {
	var tickets []Ticket
	if err := db.
		Preload("TicketOrder.TicketRelease.Event").
		Preload("TicketOrder.TicketRelease.TicketReleaseMethodDetail").
		Preload("TicketOrder.TicketRelease.PaymentDeadline").
		Preload("TicketOrder.TicketRelease.AddOns").
		Preload("TicketType").
		Preload("TicketAddOns").
		Where("user_ug_kth_id = ?", userUGKthID).
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetAllTicketsToTicketRelease(db *gorm.DB, ticketReleaseID uint) (tickets []Ticket, err error) {
	// Get all tickets to a ticket release thats not soft deleted or reserved
	err = db.
		Preload("TicketType").
		Preload("TicketOrder.TicketRelease.Event.Organization").
		Joins("JOIN ticket_orders ON tickets.ticket_order_id = ticket_orders.id").
		Joins("JOIN ticket_releases ON ticket_orders.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.id = ? AND tickets.refunded = ? AND tickets.is_reserve = ?", ticketReleaseID, false, false).
		Find(&tickets).Error
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetAllReserveTicketsToTicketRelease(db *gorm.DB, ticketReleaseID uint) (tickets []Ticket, err error) {
	// Get all tickets to a ticket release thats not soft deleted or reserved or refunded
	err = db.
		Preload("TicketRequest.User").
		Joins("JOIN ticket_requests ON tickets.ticket_request_id = ticket_requests.id").
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.id = ? AND tickets.is_reserve = ?", ticketReleaseID, false, true).
		Order("reserve_number ASC").
		Find(&tickets).Error

	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (t *Ticket) ValidatePaymentDeadline(db *gorm.DB) (err error) {
	var event Event
	db.First(&event, t.TicketOrder.TicketRelease.EventID)

	// Validate that t.TicketRequest.TicketRelease.Event.Date is loaded
	if event.Date.IsZero() {
		return fmt.Errorf("event date is not loaded")
	}

	if t.PaymentDeadline.Valid {
		if time.Now().After(t.PaymentDeadline.Time) {
			return fmt.Errorf("payment deadline has passed")
		}
	} else {
		// Check if time is after event start
		if time.Now().After(event.Date) {
			return fmt.Errorf("event has already started")
		}
	}

	return nil
}
