package models

import (
	"fmt"

	"gorm.io/gorm"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
)

// Transaction
type Transaction struct {
	gorm.Model
	PaymentIntentID   string             `json:"payment_intent_id"`
	EventID           int                `json:"event_id"`
	TicketID          int                `json:"ticket_id" gorm:"unique"`
	UserUGKthID       string             `json:"user_ug_kth_id"`
	User              User               `json:"user"`
	Amount            int                `json:"amount"`
	Currency          string             `json:"currency"`
	PayedAt           *int64             `json:"payed_at"`
	Refunded          bool               `json:"refunded" default:"false"`
	RefundedAt        *int64             `json:"refunded_at"`
	Status            TransactionStatus  `json:"status"`
	EventSalesReports []EventSalesReport `gorm:"many2many:event_sales_report_transactions;"`
}

// Validate
func (trans *Transaction) Validate() error {
	switch trans.Status {
	case TransactionStatusPending, TransactionStatusCompleted:
		return nil
	default:
		err := fmt.Errorf("invalid transaction status: %s", trans.Status)
		return err
	}
}
