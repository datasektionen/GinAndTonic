package models

import (
	"time"

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
	TicketAddOns      []TicketAddOn            `gorm:"foreignKey:TicketRequestID" json:"ticket_add_ons"`
	HandledAt         *time.Time               `json:"handled_at" gorm:"default:null"`
}

func (tr *TicketRequest) BeforeSave(tx *gorm.DB) (err error) {
	if tr.IsHandled && tr.HandledAt == nil {
		now := time.Now()
		tr.HandledAt = &now
	}
	return
}

func GetAllValidTicketRequestsToTicketRelease(db *gorm.DB, ticketReleaseID uint) ([]TicketRequest, error) {
	var ticketRequests []TicketRequest
	if err := db.
		Preload("TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("TicketRelease.PaymentDeadline").
		Preload("TicketAddOns.AddOn").
		Where("ticket_release_id = ? AND is_handled = ?", ticketReleaseID, false).Find(&ticketRequests).Error; err != nil {
		return nil, err
	}

	// According to gorm soft delete, we should not fetch soft deleted records

	return ticketRequests, nil
}

func GetValidTicketReqeust(db *gorm.DB, ticketRequestID uint) (*TicketRequest, error) {
	var ticketRequest TicketRequest
	if err := db.
		Preload("TicketType").
		Preload("TicketRelease.Event.FormFields").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("EventFormReponses").
		Where("id = ?", ticketRequestID).First(&ticketRequest).Error; err != nil {
		return nil, err
	}

	return &ticketRequest, nil
}

func GetAllValidTicketRequestToTicketReleaseOrderedByCreatedAt(db *gorm.DB, ticketReleaseID uint) ([]TicketRequest, error) {
	var ticketRequests []TicketRequest
	if err := db.
		Preload("TicketType").
		Preload("TicketRelease.Event").
		Preload("TicketRelease.PaymentDeadline").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("TicketAddOns.AddOn").
		Where("ticket_release_id = ? AND is_handled = ?", ticketReleaseID, false).Order("created_at").Find(&ticketRequests).Error; err != nil {
		return nil, err
	}

	// According to gorm soft delete, we should not fetch soft deleted records

	return ticketRequests, nil
}

func GetAllValidUsersTicketRequests(db *gorm.DB, userUGKthID string, ids *[]int) ([]TicketRequest, error) {
	var ticketRequests []TicketRequest

	query := db.
		Unscoped().
		Preload("TicketType").
		Preload("TicketRelease.Event.FormFields").
		Preload("TicketRelease.AddOns").
		Preload("TicketRelease.PaymentDeadline").
		Preload("TicketRelease.TicketReleaseMethodDetail").
		Preload("EventFormReponses").
		Preload("TicketAddOns.AddOn").
		Where("user_ug_kth_id = ?", userUGKthID).
		Find(&ticketRequests)

	if ids != nil {
		if len(*ids) > 0 {
			query = query.Where("id IN (?)", *ids)
		}
	}

	if err := query.Find(&ticketRequests).Error; err != nil {
		return nil, err
	}

	return ticketRequests, nil
}
