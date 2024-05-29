package surfboard_service_store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type UpdateStoreRequest struct {
	StoreName   string           `json:"storeName,omitempty"`
	Email       string           `json:"email,omitempty"`
	PhoneNumber *Phone           `json:"phoneNumber,omitempty"`
	Address     string           `json:"address,omitempty"`
	CareOf      string           `json:"careOf,omitempty"`
	City        string           `json:"city,omitempty"`
	ZipCode     string           `json:"zipCode,omitempty"`
	Country     string           `json:"country,omitempty"`
	OnlineInfo  *StoreOnlineInfo `json:"onlineInfo,omitempty"`
}

type StoreOnlineInfo struct {
	MerchantWebshopURL    string `json:"merchantWebshopURL,omitempty"`
	PaymentPageHostURL    string `json:"paymentPageHostURL,omitempty"`
	TermsAndConditionsURL string `json:"termsAndConditionsURL,omitempty"`
	PrivacyPolicyURL      string `json:"privacyPolicyURL,omitempty"`
}

type Phone struct {
	Code   int    `json:"code"`
	Number string `json:"number"`
}

type UpdateStoreResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func UpdateStore(merchantId string, storeId string, updateData UpdateStoreRequest) error {
	service := NewStoreService()

	updateStore := UpdateStoreRequest{
		OnlineInfo: &StoreOnlineInfo{
			MerchantWebshopURL:    "https://example.com",
			PaymentPageHostURL:    "https://example.com",
			TermsAndConditionsURL: "https://example.com",
			PrivacyPolicyURL:      "https://example.com",
		},
	}

	updateStoreBytes, err := json.Marshal(updateStore)
	if err != nil {
		return err
	}

	response, err := service.UpdateStore(merchantId, storeId, updateStoreBytes)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))

	var resp UpdateStoreResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	if resp.Status != "SUCCESS" {
		return errors.New(resp.Message)
	}

	return nil
}
