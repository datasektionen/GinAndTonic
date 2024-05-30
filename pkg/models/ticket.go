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
	Reserved  TicketStatus = "reserved"
	Cancelled TicketStatus = "cancelled"
	Allocated TicketStatus = "allocated"
)

type Ticket struct {
	gorm.Model
	TicketRequestID uint          `gorm:"index" json:"ticket_request_id"`
	TicketRequest   TicketRequest `json:"ticket_request"`
	IsPaid          bool          `json:"is_paid" default:"false"`
	IsReserve       bool          `json:"is_reserve"`
	WasReserve      bool          `json:"was_reserve" default:"false"`
	ReserveNumber   uint          `json:"reserve_number" default:"0"`
	Refunded        bool          `json:"refunded" default:"false"`
	UserUGKthID     string        `json:"user_ug_kth_id"`
	User            User          `json:"user"`
	Transaction     *Transaction  `json:"transaction"`
	Status          TicketStatus  `json:"status" gorm:"default:'pending'"`
	CheckedIn       bool          `json:"checked_in" default:"false"`
	CheckedInAt     sql.NullTime  `json:"checked_in_at"`
	QrCode          string        `json:"qr_code" gorm:"unique;not null"`
	PurchasableAt   *time.Time    `json:"purchasable_at" gorm:"default:null"`
	PaymentDeadline *time.Time    `json:"payment_deadline" gorm:"default:null"`
	TicketAddOns    []TicketAddOn `gorm:"foreignKey:TicketID" json:"ticket_add_ons"`
	DeletedReason   string        `json:"deleted_reason" gorm:"default:null"`

	OrderID *string `json:"order_id" gorm:"default:null"`
	Order   Order   `json:"order"`
}

func (t *Ticket) BeforeSave(tx *gorm.DB) (err error) {
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
	if err := db.Model(&t.TicketRequest).Update("deleted_reason", reason).Error; err != nil {
		return err
	}

	// Delete the ticket request as well
	if err := db.Delete(&t.TicketRequest).Error; err != nil {
		return err
	}

	return db.Delete(t).Error
}

func GetTicketByID(db *gorm.DB, ticketID uint) (ticket Ticket, err error) {
	err = db.
		Preload("TicketRequest.TicketType").
		First(&ticket, ticketID).Error
	if err != nil {
		return Ticket{}, err
	}

	return ticket, nil
}

func GetTicketsByIDs(db *gorm.DB, ticketIDs []uint) (tickets []Ticket, err error) {
	err = db.
		Preload("TicketRequest.TicketType").
		Preload("TicketRequest.TicketRelease").
		Find(&tickets, ticketIDs).Error
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetTicketRequestsToEvent(db *gorm.DB, eventID uint) (ticketRequests []TicketRequest, err error) {
	err = db.
		Joins("INNER JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ?", eventID).
		Find(&ticketRequests).Error

	if err != nil {
		return nil, err
	}

	return ticketRequests, nil
}

func GetTicketsToEvent(db *gorm.DB, eventID uint) (tickets []Ticket, err error) {
	// eventID is fetched in TicketRequest.TicketRelease.EventID
	err = db.
		Joins("JOIN ticket_requests ON tickets.ticket_request_id = ticket_requests.id").
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ?", eventID).
		Find(&tickets).Error

	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetAllPaidTicketsToEvent(db *gorm.DB, eventID uint) (tickets []Ticket, err error) {
	// eventID is fetched in TicketRequest.TicketRelease.EventID
	err = db.
		Joins("JOIN ticket_requests ON tickets.ticket_request_id = ticket_requests.id").
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.event_id = ? AND tickets.is_paid = ?", eventID, true).
		Find(&tickets).Error

	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetTicketToEvent(db *gorm.DB, eventID, ticketID uint) (ticket Ticket, err error) {
	// eventID is fetched in TicketRequest.TicketRelease.EventID
	err = db.
		Joins("JOIN ticket_requests ON tickets.ticket_request_id = ticket_requests.id").
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
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
		Preload("TicketRequest.TicketRelease.Event").
		Preload("TicketRequest.TicketType").
		Preload("TicketRequest.TicketRelease.TicketReleaseMethodDetail").
		Preload("TicketRequest.TicketRelease.PaymentDeadline").
		Preload("TicketAddOns").
		Preload("TicketRequest.TicketRelease.AddOns").
		Where("user_ug_kth_id = ?", userUGKthID).
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}

func GetAllTicketsToTicketRelease(db *gorm.DB, ticketReleaseID uint) (tickets []Ticket, err error) {
	// Get all tickets to a ticket release thats not soft deleted or reserved
	err = db.
		Preload("TicketRequest.TicketType").
		Preload("TicketRequest.TicketRelease.Event.Organization").
		Joins("JOIN ticket_requests ON tickets.ticket_request_id = ticket_requests.id").
		Joins("JOIN ticket_releases ON ticket_requests.ticket_release_id = ticket_releases.id").
		Where("ticket_releases.id = ? AND tickets.is_reserve = ?", ticketReleaseID, false).
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
		Where("ticket_releases.id = ? AND tickets.refunded = ? AND tickets.is_reserve = ?", ticketReleaseID, false, true).
		Order("reserve_number ASC").
		Find(&tickets).Error

	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (t *Ticket) ValidatePaymentDeadline() (err error) {
	// Validate that t.TicketRequest.TicketRelease.Event.Date is loaded
	if t.TicketRequest.TicketRelease.Event.ID == 0 {
		// Throw error
		return fmt.Errorf("event not loaded")
	}

	if t.PaymentDeadline != nil {
		if time.Now().After(*t.PaymentDeadline) {
			return fmt.Errorf("payment deadline has passed")
		}
	} else {
		// Check if time is after event start
		if time.Now().After(t.TicketRequest.TicketRelease.Event.Date) {
			return fmt.Errorf("event has already started")
		}
	}

	return nil
}
