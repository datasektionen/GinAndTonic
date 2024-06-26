package models

import (
	"github.com/go-playground/validator"
	"gorm.io/gorm"
)

type AddOn struct {
	gorm.Model
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Price           float64       `json:"price" validate:"gte=0"`
	MaxQuantity     int           `json:"max_quantity" validate:"gte=0"`
	ContainsAlcohol bool          `json:"contains_alcohol" gorm:"default:false"`
	IsEnabled       bool          `json:"is_enabled"`
	TicketReleaseID int           `json:"ticket_release_id"`
	TicketRelease   TicketRelease `json:"-"`
}

func (a *AddOn) ValidateAddOn() error {
	validate := validator.New()
	return validate.Struct(a)
}
