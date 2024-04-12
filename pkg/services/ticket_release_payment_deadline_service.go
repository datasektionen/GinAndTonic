package services

import (
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type TicketReleasePaymentDeadline struct {
	DB *gorm.DB
}

func NewTicketReleasePaymentDeadline(db *gorm.DB) *TicketReleasePaymentDeadline {
	return &TicketReleasePaymentDeadline{DB: db}
}

func (trpd *TicketReleasePaymentDeadline) UpdatePaymentDeadline(ticketReleaseID int, body types.PaymentDeadlineRequest) (rerr *types.ErrorResponse) {
	// Get ticket release payment deadline
	var ticketRelease models.TicketRelease
	if err := trpd.DB.
		Preload("PaymentDeadline").
		Preload("Event").Where("id = ?", ticketReleaseID).First(&ticketRelease).Error; err != nil {
		return &types.ErrorResponse{Message: "Invalid ticket release ID", StatusCode: 400}
	}

	// If open or not open ye
	if ticketRelease.IsOpen() {
		return &types.ErrorResponse{Message: "Ticket release is open", StatusCode: 400}
	}
	if ticketRelease.HasNotOpenedYet() {
		return &types.ErrorResponse{Message: "Ticket release has not opened yet", StatusCode: 400}
	}

	duration, err := time.ParseDuration(body.ReservePaymentDuration)
	if err != nil {
		return &types.ErrorResponse{Message: "Invalid duration", StatusCode: 400}
	}

	// Update payment deadline
	paymentDeadline := ticketRelease.PaymentDeadline

	// If after date if end date is not set or after end date
	if ticketRelease.Event.EndDate != nil && body.OriginalDeadline.After(*ticketRelease.Event.EndDate) {
		return &types.ErrorResponse{Message: "Original deadline is after event end date", StatusCode: 400}
	} else if ticketRelease.Event.EndDate == nil && body.OriginalDeadline.After(ticketRelease.Event.Date) {
		return &types.ErrorResponse{Message: "Original deadline is after event start date", StatusCode: 400}
	}

	paymentDeadline.OriginalDeadline = body.OriginalDeadline
	paymentDeadline.ReservePaymentDuration = &duration

	// Save payment deadline
	if err := trpd.DB.Save(&paymentDeadline).Error; err != nil {
		return &types.ErrorResponse{Message: "Failed to update payment deadline", StatusCode: 500}
	}

	return nil
}
