package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type TicketReleaseMethodDetail struct {
	gorm.Model
	MaxTicketsPerUser  uint   `json:"max_tickets_per_user"`
	NotificationMethod string `json:"notification_method"`
	CancellationPolicy string `json:"cancellation_policy"`

	OpenWindowDuration int64 `gorm:"open_window_duration" json:"open_window_duration"` // Specific to FCFS_Lottery

	TicketReleaseMethodID uint                `json:"ticket_release_method_id"`
	TicketReleaseMethod   TicketReleaseMethod `json:"ticket_release_method"`
}

// Create enum of cancellation policies
const (
	FULL_REFUND = "FULL_REFUND"
	NO_REFUND   = "NO_REFUND"
)

// Notification methods
const (
	EMAIL = "EMAIL"
)

type TicketReleaseConfig interface {
	Validate() error
}

type FCFSLotteryConfig struct {
	OpenWindowDuration int64 // In seconds
}

func (f *FCFSLotteryConfig) Validate() error {
	if f.OpenWindowDuration <= 0 {
		return errors.New("open window duration must be greater than 0")
	}
	return nil
}

func (trmd *TicketReleaseMethodDetail) ValidateCancellationPolicy() error {
	switch trmd.CancellationPolicy {
	case FULL_REFUND, NO_REFUND:
		return nil
	default:
		return fmt.Errorf("invalid CancellationPolicy: %v", trmd.CancellationPolicy)
	}
}

func (trmd *TicketReleaseMethodDetail) ValidateNotificationMethod() error {
	switch trmd.NotificationMethod {
	case EMAIL:
		return nil
	default:
		return fmt.Errorf("invalid NotificationMethod: %v", trmd.NotificationMethod)
	}
}

func (trmd *TicketReleaseMethodDetail) ValidateMaxTicketsPerUser() error {
	if trmd.MaxTicketsPerUser <= 0 {
		return fmt.Errorf("MaxTicketsPerUser must be greater than zero")
	}

	if trmd.MaxTicketsPerUser > 10 {
		return fmt.Errorf("MaxTicketsPerUser must be less than or equal to 10")
	}

	return nil
}

// Validate the ticket release method detail
func (trmd *TicketReleaseMethodDetail) Validate() error {
	if err := trmd.ValidateCancellationPolicy(); err != nil {
		return err
	}

	if err := trmd.ValidateNotificationMethod(); err != nil {
		return err
	}

	if err := trmd.ValidateMaxTicketsPerUser(); err != nil {
		return err
	}

	return nil
}

func NewTicketReleaseConfig(methodName string, detail *TicketReleaseMethodDetail) (TicketReleaseConfig, error) {
	switch methodName {
	case string(FCFS_LOTTERY):
		return &FCFSLotteryConfig{
			OpenWindowDuration: detail.OpenWindowDuration,
		}, nil
	default:
		return nil, fmt.Errorf("unknown method: %s", methodName)
	}
}
