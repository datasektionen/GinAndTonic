package surfboard_service_order

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func CreateOrder() {
	var jsonStr = []byte(`{
        "terminal$id": "813bee989f08500405",
        "type": "purchase",
        "orderLines": [
            {
                "id": "1234",
                "name": "Bucket hat",
                "quantity": 1,
                "itemAmount": {
                    "regular": 2000,
                    "total": 2000,
                    "currency": "SEK",
                    "tax": [
                        {
                            "amount": 200,
                            "percentage": 10,
                            "type": "vat"
                        }
                    ]
                }
            }
        ]
    }`)

	req, err := http.NewRequest(
		http.MethodPost,
		"YOUR_API_URL/orders",
		bytes.NewBuffer(jsonStr),
	)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("API-KEY", "YOUR_API_KEY")
	req.Header.Add("API-SECRET", "YOUR_API_SECRET")
	req.Header.Add("MERCHANT-ID", "YOUR_MERCHANT_ID")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
