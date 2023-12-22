package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	TicketID    uint       `json:"ticket_id"`
	Ticket      Ticket     `json:"ticket"`
	UserUGKthID string     `json:"user_ug_kth_id"`
	User        User       `json:"user"`
	Amount      int        `json:"amount"`
	Currency    string     `json:"currency"`
	PayedAt     time.Time  `json:"payed_at"`
	Refunded    bool       `json:"refunded" default:"false"`
	RefundedAt  *time.Time `json:"refunded_at"`
}
