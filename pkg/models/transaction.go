package models

import (
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	TicketID    uint   `json:"ticket_id"`
	UserUGKthID string `json:"user_ug_kth_id"`
	User        User   `json:"user"`
	Amount      int    `json:"amount"`
	Currency    string `json:"currency"`
	PayedAt     int64  `json:"payed_at"`
	Refunded    bool   `json:"refunded" default:"false"`
	RefundedAt  *int64 `json:"refunded_at"`
}
