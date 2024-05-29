package models

import (
	"time"

	"gorm.io/gorm"
)

type OnlineTerminalModeType string

const (
	OnlineTerminalPaymentPage OnlineTerminalModeType = "PaymentPage"
)

/*
A terminal is created when a new event is created. The terminal is used in
combination with the organizations store to create an online
terminal for the event. The terminal is used for guests to pay for tickets.
*/
type StoreTerminal struct {
	TerminalID string `json:"terminal_id" gorm:"primaryKey"`
	EventID    uint   `json:"event_id"`
	StoreID    string `json:"store_id"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
