package surfboard_service_order

import (
	"fmt"
	"net/http"

	surfboard_service "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard"
	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
)

type OrderEndpoint string

const (
	CreateOrderEndpoint    OrderEndpoint = "/orders"
	GetOrderStatusEndpoint OrderEndpoint = "/orders/%s/status"
)

type OrderService struct {
	client surfboard_service.SurfboardClient
}

func NewOrderService() *OrderService {
	return &OrderService{client: surfboard_service.NewSurfboardClient()}
}

func (s *OrderService) createOrder(merchantId string, data []byte) (*http.Response, error) {
	endpoint := fmt.Sprintf(string(CreateOrderEndpoint))
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:     http.MethodPost,
		Endpoint:   endpoint,
		JSONStr:    &data,
		MerchantId: &merchantId,
	})
}

func (s *OrderService) getOrderStatus(merchantId, orderId string) (*http.Response, error) {
	endpoint := fmt.Sprintf(string(GetOrderStatusEndpoint), orderId)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:     http.MethodGet,
		Endpoint:   endpoint,
		MerchantId: &merchantId,
	})
}
