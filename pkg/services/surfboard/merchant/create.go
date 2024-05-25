/*
Creates a new merchant
*/

package surfboard_service_merchant

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type Address struct {
	CareOf       string `json:"careOf"`
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2,omitempty"`
	City         string `json:"city"`
	CountryCode  string `json:"countryCode"`
	PostalCode   string `json:"postalCode"`
}

type Phone struct {
	Code   int    `json:"code"`
	Number string `json:"number"`
}

type Organisation struct {
	CorporateID string  `json:"corporateId"`
	LegalName   string  `json:"legalName"`
	MccCode     string  `json:"mccCode,omitempty"`
	Address     Address `json:"address"`
	Phone       Phone   `json:"phone"`
	Email       string  `json:"email"`
}

type DisplayProduct struct {
	ProductID    string   `json:"productId"`
	PricingPlans []string `json:"pricingPlans"`
}

type PreSelectProduct struct {
	ProductID     string `json:"productId"`
	Quantity      int    `json:"quantity"`
	PricingPlanID string `json:"pricingPlanId"`
}

type Store struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	PhoneNumber Phone   `json:"phoneNumber"`
	Address     Address `json:"address"`
}

type ControlFields struct {
	DisplayProducts        []DisplayProduct   `json:"displayProducts"`
	ShowProductCatalogue   bool               `json:"showProductCatalogue"`
	TransactionPricingPlan string             `json:"transactionPricingPlan"`
	PreSelectProducts      []PreSelectProduct `json:"preSelectProducts"`
	Store                  Store              `json:"store"`
}

type CreateMerchantType struct {
	Country         string        `json:"country"`
	Organisation    Organisation  `json:"organisation"`
	AcquirerMID     string        `json:"acquirerMID,omitempty"`
	MultiMerchantID string        `json:"multiMerchantId,omitempty"`
	ControlFields   ControlFields `json:"controlFields"`
}

type CreateMerchantResponse struct {
	Status string `json:"status"`
	Data   struct {
		ApplicationID     string                       `json:"applicationId"`
		WebKybURL         string                       `json:"webKybUrl,omitempty"`
		ApplicationStatus models.SurfApplicationStatus `json:"applicationStatus,omitempty"`
		MerchantID        string                       `json:"merchantId,omitempty"`
		StoreID           string                       `json:"storeId,omitempty"`
	} `json:"data"`
	Message string `json:"message"`
}

func CreateMerchant(tx *gorm.DB, user *models.User, network *models.Network) error {
	service := NewMerchantService()

	details := network.Details

	merchant := CreateMerchantType{
		Country: details.CountryCode,
		Organisation: Organisation{
			CorporateID: details.CorporateID,
			LegalName:   details.LegalName,
			Address: Address{
				AddressLine1: details.AddressLine1,
				City:         details.City,
				CountryCode:  details.CountryCode,
				PostalCode:   details.PostalCode,
			},
			Phone: Phone{
				Code:   details.PhoneCode,
				Number: details.PhoneNumber,
			},
		},
		ControlFields: ControlFields{
			Store: Store{
				Name:  details.StoreName,
				Email: details.Email,
				PhoneNumber: Phone{
					Code:   details.PhoneCode,
					Number: details.PhoneNumber,
				},
				Address: Address{
					AddressLine1: details.AddressLine1,
					City:         details.City,
					CountryCode:  details.CountryCode,
					PostalCode:   details.PostalCode,
				},
			},
		},
	}

	merchantBytes, err := json.Marshal(merchant)
	if err != nil {
		fmt.Println(err)
		return err
	}

	response, err := service.CreateMerchant(merchantBytes)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(response.Body)

	var resp CreateMerchantResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	if resp.Status != "SUCCESS" {
		return errors.New(resp.Message)
	}

	var networkMerchant models.NetworkMerchant = models.NetworkMerchant{
		NetworkID:     network.ID,
		ApplicationID: resp.Data.ApplicationID,
	}

	if err := tx.Create(&networkMerchant).Error; err != nil {
		fmt.Println(err)
		return err
	}

	postCreateApplicationStatus(tx, networkMerchant)

	return nil
}
