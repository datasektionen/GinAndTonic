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

func CreateOnlineTerminal(networkMerchant *models.NetworkMerchant, storeId string, eventId uint, tx *gorm.DB) error {
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

	response, err := service.CreateOnlineTerminal(networkMerchant.MerchantID, storeId, jsonStr)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(response.Body)

	var resp CreateOnlineTerminalResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	if resp.Status != "SUCCESS" {
		return errors.New(resp.Message)
	}

	var terminal models.StoreTerminal = models.StoreTerminal{
		TerminalID: resp.Data.TerminalID,
		EventID:    eventId,
		StoreID:    storeId,
	}

	if err := tx.Create(&terminal).Error; err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
