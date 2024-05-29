package models

import (
	"time"
)

type OnlineTerminalModeType string

const (
	OnlineTerminalPaymentPage OnlineTerminalModeType = "PaymentPage"
)

type NetworkTerminal struct {
	TerminalID string `json:"terminal_id" gorm:"primaryKey"`
	StoreID    string `json:"store_id"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
