package surfboard_service_store

type CreateStoreRequest struct {
	StoreName   string          `json:"storeName,omitempty"`
	Email       string          `json:"email,omitempty"`
	PhoneNumber *Phone          `json:"phoneNumber,omitempty"`
	Address     string          `json:"address" binding:"required"`
	CareOf      string          `json:"careOf,omitempty"`
	City        string          `json:"city,omitempty"`
	ZipCode     string          `json:"zipCode,omitempty"`
	Country     string          `json:"country,omitempty"`
	AcquirerMID string          `json:"acquirerMID,omitempty"`
	OnlineInfo  StoreOnlineInfo `json:"onlineInfo,omitempty" binding:"required"`
}

type StoreResponse struct {
	Status string `json:"status"`
	Data   struct {
		StoreID    string `json:"storeId"`
		MerchantID string `json:"merchantId"`
		Name       string `json:"name"`
		Adresss    struct {
			CareOf       string `json:"careOf"`
			AddressLine1 string `json:"addressLine1"`
			AddressLine2 string `json:"addressLine2"`
			City         string `json:"city"`
			PostalCode   string `json:"postalCode"`
			CountryCode  string `json:"countryCode"`
		} `json:"address"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
		OnlineInfo struct {
			MerchantWebshopURL    string `json:"merchantWebshopURL"`
			PaymentPageHostURL    string `json:"paymentPageHostURL"`
			TermsAndConditionsURL string `json:"termsAndConditionsURL"`
			PrivacyPolicyURL      string `json:"privacyPolicyURL"`
		} `json:"onlineInfo"`
	} `json:"data"`
	Message string `json:"message"`
}
