package models

import (
	"gorm.io/gorm"
)

type TicketRequest struct {
	gorm.Model
	TicketAmount      int                      `json:"ticket_amount"`
	TicketReleaseID   uint                     `json:"ticket_release_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	TicketRelease     TicketRelease            `json:"ticket_release"`
	TicketTypeID      uint                     `json:"ticket_type_id" gorm:"index" `
	TicketType        TicketType               `json:"ticket_type"`
	UserUGKthID       string                   `json:"user_ug_kth_id"`
	User              User                     `json:"user"`
	IsHandled         bool                     `json:"is_handled" gorm:"default:false"`
	Tickets           []Ticket                 `json:"tickets"`
	EventFormReponses []EventFormFieldResponse `json:"event_form_responses"`
}

func GetAllValidTicketRequestsToTicketRelease(db *gorm.DB, ticketReleaseID uint) ([]TicketRequest, error) {
	var ticketRequests []TicketRequest
	if err := db.
		Preload("TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Where("ticket_release_id = ? AND is_handled = ?", ticketReleaseID, false).Find(&ticketRequests).Error; err != nil {
		return nil, err
	}

	// Acording to gorm soft delete, we should not fetch soft deleted records

	return ticketRequests, nil
}

func GetAllValidUsersTicketRequests(db *gorm.DB, userUGKthID string) ([]TicketRequest, error) {
	var ticketRequests []TicketRequest
	if err := db.
		Preload("TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("EventFormReponses").
		Where("user_ug_kth_id = ?", userUGKthID).
		Find(&ticketRequests).Error; err != nil {
		return nil, err
	}

	return ticketRequests, nil
}
