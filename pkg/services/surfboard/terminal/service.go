package surfboard_service_terminal

import (
	"fmt"
	"net/http"

	surfboard_service "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard"
)

type TerminalEndpoints string

const (
	CreateOnlineTerminalEndpoint TerminalEndpoints = "/merchants/%s/stores/%s/online-terminals" // Takes merchantID and storeID
)

type TerminalService struct {
	client surfboard_service.SurfboardClient
}

func NewTerminalService() *TerminalService {
	return &TerminalService{client: surfboard_service.NewSurfboardClient()}
}

func (s *TerminalService) CreateOnlineTerminal(merchantId string, storeId string, data []byte) (*http.Response, error) {
	endpoint := fmt.Sprintf(string(CreateOnlineTerminalEndpoint), merchantId, storeId)
	return s.client.MakeRequest(endpoint, http.MethodPost, data)
}
