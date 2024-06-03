package surfboard_service_payment

import (
	"fmt"
	"net/http"

	surfboard_service "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard"
	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
)

type PaymentEndpoint string

const (
	VoidPaymentEndpoint PaymentEndpoint = "/payments/%s/void"
)

type PaymentService struct {
	client surfboard_service.SurfboardClient
}

func NewPaymentService() *PaymentService {
	return &PaymentService{client: surfboard_service.NewSurfboardClient()}
}

func (s *PaymentService) voidPayment(merchantId, paymentId string) (*http.Response, error) {
	endpoint := fmt.Sprintf(string(VoidPaymentEndpoint), paymentId)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:     http.MethodPost,
		Endpoint:   endpoint,
		MerchantId: &merchantId,
	})
}
