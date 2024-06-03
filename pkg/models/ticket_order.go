package models

import (
	"time"

	"gorm.io/gorm"
)

type TicketOrder struct {
	gorm.Model
	UserUGKthID string          `json:"user_ug_kth_id"`
	User        User            `json:"user"`
	Requests    []TicketRequest `gorm:"foreignKey:TicketOrderID" json:"requests"`
	TotalAmount float64         `json:"total_amount"`
	IsPaid      bool            `json:"is_paid" gorm:"default:false"`
	PaidAt      *time.Time      `json:"paid_at" gorm:"default:null"`
}
