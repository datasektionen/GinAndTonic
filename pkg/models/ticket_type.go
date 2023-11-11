package models

import (
	"gorm.io/gorm"
)

type TicketType struct {
	gorm.Model
	EventID           uint    `gorm:"index" json:"event_id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Price             float64 `json:"price"`
	QuantityAvailable uint    `gorm:"default:0" json:"quantity_available"`
	QuantityTotal     uint    `json:"quantity_total"`
	IsReserved        bool    `json:"is_reserved"`

	TicketReleaseID uint `json:"ticket_release_id"`
}

func (tt *TicketType) BeforeCreate(tx *gorm.DB) (err error) {
	if tt.QuantityAvailable == 0 {
		tt.QuantityAvailable = uint(tt.QuantityTotal)
	}
	return
}
