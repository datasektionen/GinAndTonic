package surfboard_service_receipt

import (
	"fmt"
	"net/http"

	surfboard_service "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard"
	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
)

type ReceiptEndpoint string

const (
	EmailReceiptEndpoint ReceiptEndpoint = "/receipts/%s/email"
)

type ReceiptService struct {
	client surfboard_service.SurfboardClient
}

func NewReceiptService() *ReceiptService {
	return &ReceiptService{client: surfboard_service.NewSurfboardClient()}
}

func (s *ReceiptService) EmailReceipt(merchantId, orderID string, data []byte) (*http.Response, error) {
	endpoint := fmt.Sprintf(string(EmailReceiptEndpoint), orderID)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:     http.MethodPut,
		Endpoint:   endpoint,
		JSONStr:    &data,
		MerchantId: &merchantId,
	})
}
