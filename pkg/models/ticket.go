package models

import (
	"database/sql"
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
	ReserveNumber   uint          `json:"reserve_number" default:"0"`
	Refunded        bool          `json:"refunded" default:"false"`
	UserUGKthID     string        `json:"user_ug_kth_id"` // Maybe not needed
	User            User          `json:"user"`
	Transaction     *Transaction  `json:"transaction"`
	Status          TicketStatus  `json:"status" gorm:"default:'pending'"`
	CheckedIn       bool          `json:"checked_in" default:"false"`
	CheckedInAt     sql.NullTime  `json:"checked_in_at"`
	QrCode          string        `json:"qr_code" gorm:"unique;not null"`
	PurchasableAt   *time.Time    `json:"purchasable_at" gorm:"default:null"`
	PaymentDeadline *time.Time    `json:"payment_deadline" gorm:"default:null"`
	TicketAddOns    []TicketAddOn `gorm:"foreignKey:TicketID" json:"ticket_add_ons"`
}

func (t *Ticket) Delete(db *gorm.DB) error {
	// Delete the associated TicketRequest
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	println("Deleting ticket request with ID: ", t.TicketRequestID)

	if err := tx.Delete(&t.TicketRequest).Error; err != nil {
		return err
	}

	// Delete the Ticket
	err := tx.Delete(&t).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
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
