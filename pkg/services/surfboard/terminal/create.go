package surfboard_service_terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type SurfboardTerminal struct {
	OnlineTerminalMode models.OnlineTerminalModeType `json:"onlineTerminalMode"`
}

type CreateOnlineTerminalResponse struct {
	Status string `json:"status"`
	Data   struct {
		TerminalID string `json:"terminalId"`
	} `json:"data"`
	Message string `json:"message"`
}

func CreateOnlineTerminal(tx *gorm.DB, networkMerchant *models.NetworkMerchant, organization *models.Organization) error {
	service := NewTerminalService()
	if !networkMerchant.IsApplicationCompleted() {
		return fmt.Errorf("merchant application is not completed")
	}

	var data SurfboardTerminal = SurfboardTerminal{
		OnlineTerminalMode: "PaymentPage",
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(string(jsonStr))

	response, err := service.CreateOnlineTerminal(networkMerchant.MerchantID, networkMerchant.StoreID, jsonStr)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(response.Body)

	var resp CreateOnlineTerminalResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	fmt.Println("response Body:", string(body))
	if resp.Status != "SUCCESS" {
		return errors.New(resp.Message)
	}

	return nil
}

// Function CreateInitialTerminalsForNetwork takes the networkID
// and creates initial online terminals for all the organizations in the network.
func CreateInitialTerminalsForNetwork(tx *gorm.DB, networkID int) error {
	var network models.Network
	if err := tx.Preload("Merchant").Preload("Organizations").First(&network, networkID).Error; err != nil {
		return err
	}

	for _, org := range network.Organizations {
		if err := CreateOnlineTerminal(tx, &network.Merchant, &org); err != nil {
			return err
		}
	}

	return nil
}
