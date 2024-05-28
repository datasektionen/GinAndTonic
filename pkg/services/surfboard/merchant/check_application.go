/*
Creates a new merchant
*/

package surfboard_service_merchant

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	surfboard_service_terminal "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/terminal"
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

	fmt.Println(string(body))

	var resp CreateMerchantResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	if resp.Status != "SUCCESS" {
		return fmt.Errorf(resp.Message)
	}

	// Check if the application is merchant_created and the current application status is not the same
	if networkMerchant.ApplicationStatus == models.MERCHANT_CREATED && resp.Data.ApplicationStatus != models.MERCHANT_CREATED {
		err := surfboard_service_terminal.CreateInitialTerminalsForNetwork(tx, int(networkMerchant.NetworkID))
		if err != nil {
			return err
		}
	}

	var appStatus models.SurfApplicationStatus = models.SurfApplicationStatus(strings.ToLower(string(resp.Data.ApplicationStatus)))

	networkMerchant.ApplicationStatus = appStatus
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

	var resp CreateMerchantResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
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
