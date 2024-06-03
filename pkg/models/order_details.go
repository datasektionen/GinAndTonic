package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderDetails struct {
	gorm.Model
	OrderID string `json:"order_id"`

	PaymentID                  string          `json:"payment_id"`
	TransactionID              string          `json:"transaction_id"`
	PaymentMethod              string          `json:"payment_method"`
	PaymentStatus              OrderStatusType `json:"payment_status"`
	TruncatedPan               string          `json:"truncated_pan"`
	CardLabel                  string          `json:"card_label"`
	PosEntryMode               string          `json:"pos_entry_mode"`
	IssuerApplication          string          `json:"issuer_application"`
	TerminalVerificationResult string          `json:"terminal_verification_result"`
	Aid                        string          `json:"aid"`
	CustomerResponseCode       string          `json:"customer_response_code"`
	CvmMethod                  string          `json:"cvm_method"`
	AuthMode                   string          `json:"auth_mode"`

	Total    float64 `json:"total"`
	Currency string  `json:"currency"`

	Voided bool `json:"voided" gorm:"default:false"`

	Refunded   bool       `json:"refunded" gorm:"default:false"`
	RefundedAt *time.Time `json:"refunded_at"`
	PayedAt    *time.Time `json:"payed_at"`
}
