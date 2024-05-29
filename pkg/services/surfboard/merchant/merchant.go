package surfboard_service_merchant

import (
	"fmt"
	"net/http"
	"os"

	surfboard_service "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard"
	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
)

type MerchantEndpoints string

const (
	CreateMerchantEndpoint               MerchantEndpoints = "/partners/%s/merchants"
	CheckApplicationStatusEndpoint       MerchantEndpoints = "/partners/%s/merchants/%s/status"
	FetchAllMerchantApplicationsEndpoint MerchantEndpoints = "/partners/%s/merchants/applications"
	UpdateMerchantDetailsEndpoint        MerchantEndpoints = "/partners/%s/merchants/%s"
)

type MerchantService struct {
	client surfboard_service.SurfboardClient
}

func NewMerchantService() *MerchantService {
	return &MerchantService{client: surfboard_service.NewSurfboardClient()}
}

func (s *MerchantService) CreateMerchant(data []byte) (*http.Response, error) {
	SURFBOARD_PARTNER_ID := os.Getenv("SURFBOARD_PARTNER_ID")
	endpoint := fmt.Sprintf(string(CreateMerchantEndpoint), SURFBOARD_PARTNER_ID)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:    http.MethodPost,
		Endpoint:  endpoint,
		JSONStr:   &data,
		PartnerId: &SURFBOARD_PARTNER_ID,
	})
}

func (s *MerchantService) CheckApplicationStatus(applicationID string) (*http.Response, error) {
	SURFBOARD_PARTNER_ID := os.Getenv("SURFBOARD_PARTNER_ID")
	endpoint := fmt.Sprintf(string(CheckApplicationStatusEndpoint), SURFBOARD_PARTNER_ID, applicationID)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:    http.MethodGet,
		Endpoint:  endpoint,
		PartnerId: &SURFBOARD_PARTNER_ID,
	})
}

func (s *MerchantService) FetchAllMerchantApplications() (*http.Response, error) {
	SURFBOARD_PARTNER_ID := os.Getenv("SURFBOARD_PARTNER_ID")
	endpoint := fmt.Sprintf(string(FetchAllMerchantApplicationsEndpoint), SURFBOARD_PARTNER_ID)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:    http.MethodGet,
		Endpoint:  endpoint,
		PartnerId: &SURFBOARD_PARTNER_ID,
	})
}

func (s *MerchantService) UpdateMerchantDetails(merchantID string, data []byte) (*http.Response, error) {
	SURFBOARD_PARTNER_ID := os.Getenv("SURFBOARD_PARTNER_ID")
	endpoint := fmt.Sprintf(string(UpdateMerchantDetailsEndpoint), SURFBOARD_PARTNER_ID, merchantID)
	return s.client.MakeRequest(surfboard_types.SurfboardRequestArgs{
		Method:    http.MethodPost,
		Endpoint:  endpoint,
		PartnerId: &SURFBOARD_PARTNER_ID,
	})
}
