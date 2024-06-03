package surfboard_service_payment

import (
	"encoding/json"
	"errors"
	"io"
)

type VoidPaymentResponse struct {
	Status string `json:"status"`
	Data   struct {
		VoidStatus string `json:"voidStatus"`
	} `json:"data"`
	Message string `json:"message"`
}

func VoidPayment(merchantId, paymentId string) error {
	service := NewPaymentService()

	response, err := service.voidPayment(merchantId, paymentId)

	if err != nil {
		return err
	}

	body, _ := io.ReadAll(response.Body)

	var resp VoidPaymentResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	if resp.Status != "SUCCESS" {
		return errors.New(resp.Message)
	}

	if resp.Data.VoidStatus != "CANNOT_VOID" {
		return errors.New("payment cannot be voided, status is " + resp.Data.VoidStatus)
	}

	return nil
}
