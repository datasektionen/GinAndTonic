package models

import "gorm.io/gorm"

type OnlineTerminalModeType string

const (
	OnlineTerminalPaymentPage OnlineTerminalModeType = "PaymentPage"
)

type NetworkMerchantTerminals struct {
	gorm.Model
	TerminalID     string                 `json:"terminal_id" gorm:"primaryKey"`
	OrganizationID uint                   `json:"organization_id"`
	MerchantID     string                 `json:"merchant_id"`
	Type           OnlineTerminalModeType `json:"type"`
}

func (NetworkMerchantTerminals) TableName() string {
	return "network_merchant_terminals"
}
