package models

import (
	"github.com/go-playground/validator"
	"gorm.io/gorm"
)

type AddOn struct {
	gorm.Model
	Name            string        `json:"name" gorm:"unique"`
	Price           float64       `json:"price" validate:"gte=0"`
	MaxQuantity     int           `json:"max_quantity" validate:"gte=0"`
	MinQuantity     int           `json:"min_quantity" validate:"gte=0,ltefield=MaxQuantity"`
	IsEnabled       bool          `json:"is_enabled" gorm:"default:true"`
	TicketReleaseID int           `json:"ticket_release_id"`
	TicketRelease   TicketRelease `json:"ticket_release"`
}

func (a *AddOn) ValidateAddOn() error {
	validate := validator.New()
	return validate.Struct(a)
}
