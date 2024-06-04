package models

import "gorm.io/gorm"

type TicketAddOn struct {
	gorm.Model
	AddOnID  uint  `json:"add_on_id"`
	AddOn    AddOn `json:"add_on"`
	TicketID *uint `json:"ticket_id"`
	Quantity int   `json:"quantity"`
}
