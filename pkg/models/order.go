package models

import "gorm.io/gorm"

type OrderStatus string

const (
	OrderStatusPending                 OrderStatus = "pending"
	OrderStatusInitiated               OrderStatus = "payment_initiated"
	OrderStatusProcessed               OrderStatus = "payment_processed"
	OrderStatusPaymentCompleted        OrderStatus = "payment_completed"
	OrderStatusPaymentCancelled        OrderStatus = "payment_cancelled"
	OrderStatusPartialPaymentCompleted OrderStatus = "partial_payment_completed"
)

type Order struct {
	gorm.Model
	OrderID         string `json:"orderId"`
	MerchantID      string `json:"merchantId"`
	EventID         uint   `json:"event_id"`
	UserUGKthID     string `json:"user_ug_kth_id"`
	PaymentPageLink string `json:"paymentPageLink"`

	Status  OrderStatus `json:"status" gorm:"type:varchar(255);default:'pending'"`
	Tickets []Ticket    `json:"tickets" gorm:"foreignKey:OrderID"`
}

func (o Order) IsPaymentCompleted() bool {
	return o.Status == OrderStatusPaymentCompleted
}

func GetAllIncompleteOrders(db *gorm.DB) ([]Order, error) {
	var orders []Order
	if err := db.Where("(status != ? AND status != ?) or status is null", OrderStatusPaymentCompleted, OrderStatusPaymentCancelled).Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}
