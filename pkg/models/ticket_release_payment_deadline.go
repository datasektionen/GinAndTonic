package models

import (
	"fmt"
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
func (trpd *TicketReleasePaymentDeadline) Validate(ticketRelease *TicketRelease, event *Event) bool {
	if (event.ID == 0) || (event.Date.IsZero()) {
		fmt.Println("event not found, not preloaded")
		return false
	}

	if trpd.ReservePaymentDuration != nil {
		if *trpd.ReservePaymentDuration < 0 {
			return false
		}

		if event.Date.Unix() < time.Now().Add(*trpd.ReservePaymentDuration).Unix() {
			return false
		}
	}

	return true
}

func CreateReservedTicketReleasePaymentDeadline(db *gorm.DB, ticketReleaseID uint) (*TicketReleasePaymentDeadline, error) {
	// Get event date
	var ticketRelease TicketRelease
	if err := db.Preload("Event").First(&ticketRelease, ticketReleaseID).Error; err != nil {
		return nil, err
	}

	trpd := &TicketReleasePaymentDeadline{
		TicketReleaseID:        ticketReleaseID,
		OriginalDeadline:       ticketRelease.Event.Date,
		ReservePaymentDuration: nil,
	}

	if err := db.Create(trpd).Error; err != nil {
		return nil, err
	}

	return trpd, nil
}
