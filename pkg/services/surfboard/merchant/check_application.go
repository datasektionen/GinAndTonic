/*
Creates a new merchant
*/

package surfboard_service_merchant

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func CheckApplicationStatus(tx *gorm.DB, networkMerchant *models.NetworkMerchant) error {
	service := NewMerchantService()

	response, err := service.CheckApplicationStatus(networkMerchant.ApplicationID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	body, _ := io.ReadAll(response.Body)

	var resp CreateMerchantResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("Error decoding response body:", err)
		return err
	}

	if resp.Status != "SUCCESS" {
		fmt.Println("Error creating merchant:", resp.Message)
		return fmt.Errorf(resp.Message)
	}

	networkMerchant.ApplicationStatus = resp.Data.ApplicationStatus
	networkMerchant.MerchantID = resp.Data.MerchantID
	networkMerchant.WebKybURL = resp.Data.WebKybURL
	networkMerchant.StoreID = resp.Data.StoreID

	if err := tx.Save(&networkMerchant).Error; err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func postCreateApplicationStatus(tx *gorm.DB, networkMerchant models.NetworkMerchant) error {
	service := NewMerchantService()

	response, err := service.CheckApplicationStatus(networkMerchant.ApplicationID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	body, _ := io.ReadAll(response.Body)

	fmt.Println("response Body:", string(body))

	var resp CreateMerchantResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("Error decoding response body:", err)
		return err
	}

	networkMerchant.ApplicationID = resp.Data.ApplicationID
	networkMerchant.ApplicationStatus = resp.Data.ApplicationStatus
	// Will probably not be set since merchant needs to be onboarded first
	networkMerchant.MerchantID = resp.Data.MerchantID
	networkMerchant.WebKybURL = resp.Data.WebKybURL
	// Same for this
	networkMerchant.StoreID = resp.Data.StoreID

	// Save
	if err := tx.Save(&networkMerchant).Error; err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
