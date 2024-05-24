package types

type MerchantBusinessData struct {
	CountryCode   string `json:"country_code"`
	LegalName     string `json:"legal_name"`
	CorporateID   string `json:"corporate_id"`
	AddressLine1  string `json:"address_line1"`
	AddressLine2  string `json:"address_line2"`
	City          string `json:"city"`
	PostalCode    string `json:"postal_code"`
	PhoneNumber   string `json:"phone_number"`
	BusinessEmail string `json:"business_email"`
	StoreName     string `json:"store_name"`
	PhoneCode     int    `json:"phone_code"`
}
