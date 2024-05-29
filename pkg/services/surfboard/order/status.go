package surfboard_service_order

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

type OrderStatusReponse struct {
	Status string `json:"status"`
	Data   struct {
		OrderStatus string `json:"orderStatus"`
	} `json:"data"`
	Message string `json:"message"`
}

func (os *OrderService) GetOrderStatus(order *models.Order) (*string, error) {
	// Get the order status
	response, err := os.getOrderStatus(order.MerchantID, order.OrderID)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(response.Body)

	var resp OrderStatusReponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Status != "SUCCESS" {
		return nil, errors.New(resp.Message)
	}

	return &resp.Data.OrderStatus, nil
}
