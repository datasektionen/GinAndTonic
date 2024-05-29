package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	OrderID         string `json:"orderId"`
	MerchantID      string `json:"merchantId"`
	EventID         uint   `json:"event_id"`
	UserUGKthID     string `json:"user_ug_kth_id"`
	PaymentPageLink string `json:"paymentPageLink"`

	Tickets []Ticket `json:"tickets" gorm:"foreignKey:OrderID"`
}
