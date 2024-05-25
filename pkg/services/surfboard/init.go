package surfboard_service

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
)

type SurfboardClient interface {
	MakeRequest(endpoint string, method string, jsonStr []byte) (*http.Response, error)
}

type surfboardClientImpl struct {
	client *http.Client
}

func NewSurfboardClient() SurfboardClient {
	return &surfboardClientImpl{
		client: &http.Client{},
	}
}
func (s *surfboardClientImpl) MakeRequest(endpoint string, method string, jsonStr []byte) (*http.Response, error) {
	SURFBOARD_API_URL := os.Getenv("SURFBOARD_API_URL")
	SURFBOARD_API_KEY := os.Getenv("SURFBOARD_API_KEY")
	SURFBOARD_API_SECRET := os.Getenv("SURFBOARD_API_SECRET")
	// SURFBOARD_MERCHANT_ID := os.Getenv("SURFBOARD_MERCHANT_ID")
	// SURFBOARD_STORE_ID    := os.Getenv("SURFBOARD_STORE_ID")
	// SURFBOARD_PARTNER_ID  := os.Getenv("SURFBOARD_PARTNER_ID")

	req, err := http.NewRequest(
		method,
		SURFBOARD_API_URL+endpoint,
		bytes.NewBuffer(jsonStr),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("API-KEY", SURFBOARD_API_KEY)
	req.Header.Add("API-SECRET", SURFBOARD_API_SECRET)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}

	return resp, nil
}
