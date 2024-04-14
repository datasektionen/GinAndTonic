package models

import (
	"fmt"

	"github.com/stripe/stripe-go"
	"gorm.io/gorm"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
)

type TransactionType string

const (
	TypePurchase   TransactionType = "purchase"
	TypeRefund     TransactionType = "refund"
	TypeAdjustment TransactionType = "adjustment"
)

// Transaction
type Transaction struct {
	gorm.Model
	PaymentIntentID   string                    `json:"payment_intent_id"`
	EventID           int                       `json:"event_id"`
	TicketID          int                       `json:"ticket_id" gorm:"unique"`
	UserUGKthID       string                    `json:"user_ug_kth_id"`
	User              User                      `json:"user"`
	Amount            int                       `json:"amount"`
	Currency          string                    `json:"currency"`
	PayedAt           *int64                    `json:"payed_at"`
	Refunded          bool                      `json:"refunded" default:"false"`
	RefundedAt        *int64                    `json:"refunded_at"`
	Status            TransactionStatus         `json:"status"`
	PaymentMethod     *stripe.PaymentMethodType `json:"payment_method"`
	TransactionType   TransactionType           `json:"transaction_type" default:"purchase"`
	EventSalesReports []EventSalesReport        `gorm:"many2many:event_sales_report_transactions;"`
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

func GetEventTotalIncome(db *gorm.DB, eventID int) (float64, error) {
	var totalIncome float64

	err := db.Model(&Transaction{}).Where("event_id = ?", eventID).Where("status = ?", "completed").Select("SUM(amount)").Scan(&totalIncome).Error

	if err != nil {
		return 0, err
	}

	return totalIncome, nil
}
