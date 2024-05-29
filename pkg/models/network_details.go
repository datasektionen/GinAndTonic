package models

import "gorm.io/gorm"

type NetworkDetails struct {
	gorm.Model
	NetworkID    uint   `json:"network_id"`
	CorporateID  string `json:"corporate_id"` // Corporate ID of the network
	LegalName    string `json:"legal_name"`   // Legal name of the network
	Descrition   string `json:"description"`  // Description of the network
	StoreName    string `json:"store_name"`   // Store name of the network
	Language     string `json:"language"`
	CareOf       string `json:"care_of"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	PostalCode   string `json:"postal_code"`
	City         string `json:"city"`
	PhoneCode    int    `json:"phone_code"`
	PhoneNumber  string `json:"phone_number"`
	CountryCode  string `json:"country_code"` // Two-letter ISO country code, in uppercase. i.e 'SE' | 'DK' | 'FI'.
	Email        string `json:"email"`        // main email for the network
}
