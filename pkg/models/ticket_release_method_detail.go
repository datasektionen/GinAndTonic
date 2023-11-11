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

	OpenWindowDuration uint `gorm:"open_window_duration" json:"open_window_duration"` // Specific to FCFS_Lottery

	TicketReleaseMethodID uint                `json:"ticket_release_method_id"`
	TicketReleaseMethod   TicketReleaseMethod `json:"ticket_release_method"`
}

type TicketReleaseConfig interface {
	Validate() error
}

type FCFSLotteryConfig struct {
	OpenWindowDuration uint // In seconds
}

func (f *FCFSLotteryConfig) Validate() error {
	if f.OpenWindowDuration <= 0 {
		return errors.New("open window duration must be greater than 0")
	}
	return nil
}

func NewTicketReleaseConfig(methodName string, detail *TicketReleaseMethodDetail) (TicketReleaseConfig, error) {
	println(string(FCFS_LOTTERY))
	switch methodName {
	case string(FCFS_LOTTERY):
		return &FCFSLotteryConfig{
			OpenWindowDuration: detail.OpenWindowDuration,
		}, nil
	default:
		return nil, fmt.Errorf("unknown method: %s", methodName)
	}
}
