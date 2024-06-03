package surfboard_service_receipt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type ReceiptRequestBody struct {
	Email string `json:"email"`
}

type ReceiptResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func EmailReceipt(tx *gorm.DB, merchantId, orderID string) error {
	service := NewReceiptService()

	var order models.Order
	if err := tx.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return err
	}

	var user models.User
	if err := tx.Where("id = ?", order.UserUGKthID).First(&user).Error; err != nil {
		return err
	}

	var reqBody ReceiptRequestBody = ReceiptRequestBody{
		Email: user.Email,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		panic(err)
	}

	response, err := service.EmailReceipt(merchantId, orderID, reqBytes)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(response.Body)

	fmt.Println("Email receipt response: ", string(body))

	var resp ReceiptResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	fmt.Println("Email receipt response: ", resp)

	if resp.Status != "SUCCESS" {
		return errors.New(resp.Message)
	}

	return nil
}
