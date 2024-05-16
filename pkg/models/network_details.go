package models

import "gorm.io/gorm"

type NetworkDetails struct {
	gorm.Model
	NetworkID   uint   `json:"network_id"`
	CorporateID string `json:"corporate_id"` // Corporate ID of the network
	LegalName   string `json:"legal_name"`   // Legal name of the network
	Descrition  string `json:"description"`  // Description of the network
	Language    string `json:"language"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"` // Two-letter ISO country code, in uppercase. i.e 'SE' | 'DK' | 'FI'.
	Email       string `json:"email"`        // main email for the network
}
