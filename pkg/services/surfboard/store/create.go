package surfboard_service_store

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func CreateStore(network *models.Network, organization *models.Organization, tx *gorm.DB) error {
	service := NewStoreService()

	if network.Merchant.ID == 0 {
		return fmt.Errorf("merchant is not created")
	}

	if network.Details.ID == 0 {
		return fmt.Errorf("network details are not created")
	}

	networkDetails := network.Details

	// initially some of these fields are defaulted to the same as the organization,
	data := CreateStoreRequest{
		StoreName: organization.Name,
		Email:     organization.Email,
		PhoneNumber: &Phone{
			Code:   networkDetails.PhoneCode,
			Number: networkDetails.PhoneNumber,
		},
		Address: networkDetails.AddressLine1,
		CareOf:  networkDetails.CareOf,
		City:    networkDetails.City,
		ZipCode: networkDetails.PostalCode,
		Country: networkDetails.CountryCode,
		OnlineInfo: StoreOnlineInfo{
			MerchantWebshopURL:    "https://example.com",
			PaymentPageHostURL:    "https://example.com",
			TermsAndConditionsURL: "https://example.com",
			PrivacyPolicyURL:      "https://example.com",
		},
	}

	createStoreBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	response, err := service.CreateStore(network.Merchant.MerchantID, createStoreBytes)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(response.Body)

	var resp StoreResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	if resp.Status != "SUCCESS" {
		return fmt.Errorf(resp.Message)
	}

	var store models.OrganizationStore = models.OrganizationStore{
		OrganizationID: organization.ID,
		StoreID:        resp.Data.StoreID,
		Name:           resp.Data.Name,
	}

	if err := tx.Save(&store).Error; err != nil {
		return err
	}

	return nil
}
