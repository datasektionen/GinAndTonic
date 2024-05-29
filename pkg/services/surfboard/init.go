package surfboard_service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
)

type SurfboardClient interface {
	MakeRequest(args surfboard_types.SurfboardRequestArgs) (*http.Response, error)
}

type surfboardClientImpl struct {
	client    *http.Client
	baseURL   string
	apiKey    string
	apiSecret string
}

func NewSurfboardClient() SurfboardClient {
	// Initialize the client with environment variables to avoid fetching them with every request.
	return &surfboardClientImpl{
		client:    &http.Client{},
		baseURL:   os.Getenv("SURFBOARD_API_URL"),
		apiKey:    os.Getenv("SURFBOARD_API_KEY"),
		apiSecret: os.Getenv("SURFBOARD_API_SECRET"),
	}
}

func (s *surfboardClientImpl) MakeRequest(args surfboard_types.SurfboardRequestArgs) (*http.Response, error) {
	fullURL := s.baseURL + args.Endpoint
	var req *http.Request
	var err error

	if args.Method == http.MethodPost || args.Method == http.MethodPut {
		if args.JSONStr == nil {
			return nil, fmt.Errorf("JSONStr is required for POST and PUT requests")
		}

		json := *args.JSONStr
		req, err = http.NewRequest(args.Method, fullURL, bytes.NewBuffer(json))
	} else {
		req, err = http.NewRequest(args.Method, fullURL, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("API-KEY", s.apiKey)
	req.Header.Add("API-SECRET", s.apiSecret)
	if args.MerchantId != nil {
		req.Header.Add("MERCHANT-ID", *args.MerchantId)
	}
	if args.StoreId != nil {
		req.Header.Add("STORE-ID", *args.StoreId)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return nil, err
		}
		log.Printf("Error response (%d): %s", resp.StatusCode, string(bodyBytes))
		return resp, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}
