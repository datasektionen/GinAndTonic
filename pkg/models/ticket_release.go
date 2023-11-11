package models

import (
	"gorm.io/gorm"
)

type TicketRelease struct {
	gorm.Model
	EventID        int             `gorm:"index" json:"event_id"`
	Event          Event           `json:"event"`
	Open           uint            `json:"open"`
	Close          uint            `json:"close"`
	TicketTypes    []TicketType    `gorm:"foreignKey:TicketReleaseID" json:"ticket_types"`
	TicketRequests []TicketRequest `gorm:"foreignKey:TicketReleaseID" json:"ticket_requests"`

	HasAllocatedTickets bool `json:"has_allocated_tickets"`

	TicketReleaseMethodDetailID uint                      `gorm:"index" json:"ticket_release_method_detail_id"`
	TicketReleaseMethodDetail   TicketReleaseMethodDetail `json:"ticket_release_method_detail"`
}
