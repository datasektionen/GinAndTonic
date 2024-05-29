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
		return err
	}

	response, err := service.CreateMerchant(merchantBytes)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))

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
