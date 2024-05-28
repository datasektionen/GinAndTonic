package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	OrderID         string `json:"orderId"`
	Nonce           string `json:"nonce"`
	PaymentPageLink string `json:"paymentPageLink"`
}
