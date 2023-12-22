package models

import (
	"gorm.io/gorm"
)

type Ticket struct {
	gorm.Model
	TicketRequestID uint          `gorm:"index" json:"ticket_request_id"`
	TicketRequest   TicketRequest `json:"ticket_request"`
	IsPaid          bool          `json:"is_paid" default:"false"`
	IsReserve       bool          `json:"is_reserve"`
	Refunded        bool          `json:"refunded" default:"false"`
	UserUGKthID     string        `json:"user_ug_kth_id"`
	User            User          `json:"user"`
	Transaction     Transaction   `json:"transaction"`
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
		Where("user_ug_kth_id = ?", userUGKthID).
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}
