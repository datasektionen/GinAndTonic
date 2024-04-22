package models

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	gorm.Model
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`

	IsSaved                 bool       `json:"saved" gorm:"default:false"` // Used to determine if the customer wants to use this information for future events
	VerifiedEmail           *bool      `json:"verified_email" gorm:"default:false"`
	EmailVerificationToken  string     `gorm:"size:255" json:"-"`
	EmailVerificationSentAt *time.Time `json:"-"`
	PasswordHash            *string    `json:"-" gorm:"column:password_hash;default:NULL"`

	Tickets         []Ticket           `json:"tickets"`
	TicketRequests  []TicketRequest    `gorm:"foreignKey:UserUGKthID" json:"ticket_requests"`
	FoodPreferences UserFoodPreference `gorm:"foreignKey:UserUGKthID" json:"food_preferences"`
}
