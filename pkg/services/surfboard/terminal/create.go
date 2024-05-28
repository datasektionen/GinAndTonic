package surfboard_service_terminal

import (
    "bytes"
    "fmt"
    "net/http"
)


func main() {
    var jsonStr = [] byte(`
      {
          "onlineTemrinalMode":"PaymentPage"
      }`)

	req, err := http.NewRequest(
		http.MethodPost,
		"YOUR_API_URL/merchants/:merchantId/stores/:storeId/online-terminals",
		bytes.NewBuffer(jsonStr),
	)

    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("API-KEY", "YOUR_API_KEY")
    req.Header.Add("API-SECRET", "YOUR_API_SECRET")
    req.Header.Add("MERCHANT-ID", "YOUR_MERCHANT_ID")

    client: = & http.Client {}
    resp, err: = client.Do(req)
    if err != nil {
        panic(err)
    }

    fmt.Println("response Status:", resp.Status)
    fmt.Println("response Headers:", resp.Header)
    body, _: = io.ReadAll(resp.Body)
    fmt.Println("response Body:", string(body))
}
