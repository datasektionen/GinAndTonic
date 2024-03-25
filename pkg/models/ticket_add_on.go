package models

import "gorm.io/gorm"

type TicketAddOn struct {
	gorm.Model
	AddOnID         int  `json:"add_on_id"`
	TicketRequestID *int `json:"ticket_request_id"`
	TicketID        *int `json:"ticket_id"`
	Quantity        int  `json:"quantity"`
}
