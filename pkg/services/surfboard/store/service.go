package surfboard_service_store

import (
	"fmt"
	"net/http"
	"os"

	surfboard_service "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard"
	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
)

type StoreEndpoint string

const (
	CreateStoreEndpoint StoreEndpoint = "/partners/%s/merchants/%s/stores"
	UpdateStoreEndpoint StoreEndpoint = "/partners/%s/merchants/%s/stores/%s"
)

type StoreService struct {
	client surfboard_service.SurfboardClient
}

func NewStoreService() *StoreService {
	return &StoreService{client: surfboard_service.NewSurfboardClient()}
}

func (s *StoreService) CreateStore(merchantId string, data []byte) (*http.Response, error) {
	partnerID := os.Getenv("SURFBOARD_PARTNER_ID")
	endpoint := fmt.Sprintf(string(CreateStoreEndpoint), partnerID, merchantId)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:     http.MethodPost,
		Endpoint:   endpoint,
		JSONStr:    &data,
		PartnerId:  &partnerID,
		MerchantId: &merchantId,
	})
}

func (s *StoreService) UpdateStore(merchantId, storeId string, data []byte) (*http.Response, error) {
	partnerID := os.Getenv("SURFBOARD_PARTNER_ID")
	endpoint := fmt.Sprintf(string(UpdateStoreEndpoint), partnerID, merchantId, storeId)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:     http.MethodPut,
		Endpoint:   endpoint,
		JSONStr:    &data,
		PartnerId:  &partnerID,
		MerchantId: &merchantId,
		StoreId:    &storeId,
	})
}