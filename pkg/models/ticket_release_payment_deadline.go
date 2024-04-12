package models

import (
	"time"

	"gorm.io/gorm"
)

type TicketReleasePaymentDeadline struct {
	gorm.Model
	TicketReleaseID        uint           `gorm:"uniqueIndex" json:"ticket_release_id"`
	OriginalDeadline       time.Time      `gorm:"not null" json:"original_deadline"`
	ReservePaymentDuration *time.Duration `json:"reserve_payment_duration"` // represented as seconds
}

// ValidatePayWithin validates the pay within duration
func (trpd *TicketReleasePaymentDeadline) Validate(ticketRelease *TicketRelease) bool {
	if *trpd.ReservePaymentDuration < 0 {
		return false
	}

	if ticketRelease.Event.Date.Unix() < time.Now().Add(*trpd.ReservePaymentDuration).Unix() {
		return false
	}

	return true
}
