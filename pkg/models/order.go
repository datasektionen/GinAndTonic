package models

import (
	"errors"
	"time"

	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
	"gorm.io/gorm"
)

type OrderStatusType string

const (
	OrderStatusPending                 OrderStatusType = "pending"
	OrderStatusInitiated               OrderStatusType = "payment_initiated"
	OrderStatusProcessed               OrderStatusType = "payment_processed"
	OrderStatusPaymentCompleted        OrderStatusType = "payment_completed"
	OrderStatusPaymentFailed           OrderStatusType = "payment_failed"
	OrderStatusPaymentCancelled        OrderStatusType = "payment_cancelled"
	OrderStatusPartialPaymentCompleted OrderStatusType = "partial_payment_completed"
)

type Order struct {
	gorm.Model
	OrderID         string `json:"orderId"`
	MerchantID      string `json:"merchantId"`
	EventID         uint   `json:"event_id"`
	UserUGKthID     string `json:"user_ug_kth_id"`
	PaymentPageLink string `json:"paymentPageLink"`

	Details OrderDetails `json:"details" gorm:"foreignKey:OrderID"`
	Tickets []Ticket     `json:"tickets" gorm:"foreignKey:OrderID"`

	WebhookEvents []OrderWebhookEvent `json:"webhook_events" gorm:"foreignKey:OrderID"`
}

func GetOrderByID(db *gorm.DB, orderID string) (*Order, error) {
	var order Order
	if err := db.Preload("Details").Preload("Tickets").Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *Order) IsPaymentCompleted() bool {
	return o.Details.PaymentStatus == OrderStatusPaymentCompleted
}

func (o *Order) CanUpdateOrder() bool {
	if o.Details.PaymentStatus == OrderStatusPaymentCompleted {
		return false
	}

	if o.Details.PaymentStatus == OrderStatusPaymentCancelled {
		return false
	}

	return true
}

func GetAllIncompleteOrders(db *gorm.DB) ([]Order, error) {
	var orders []Order
	if err := db.Where("(status != ? AND status != ?) or status is null", OrderStatusPaymentCompleted, OrderStatusPaymentCancelled).Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *Order) CancelOrder(db *gorm.DB) error {
	if o.Details.PaymentStatus == OrderStatusPaymentCompleted {
		return errors.New("order has already been completed")
	}

	if o.Details.PaymentStatus == OrderStatusPartialPaymentCompleted {
		return errors.New("order has already been partially completed")
	}

	if o.Details.PaymentStatus == OrderStatusPaymentCancelled {
		return errors.New("order has already been cancelled")
	}

	o.Details.PaymentStatus = OrderStatusPaymentCancelled
	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Omit("Tickets").Save(o).Error; err != nil {
		return err
	}

	// Delete the order
	if err := db.Delete(&o).Error; err != nil {
		return err
	}

	return nil
}

func (o *Order) PaymentInitiated(db *gorm.DB, data surfboard_types.OrderWebhookData) error {
	o.Details.PaymentStatus = OrderStatusType(data.Data.PaymentStatus)
	o.Details.PaymentMethod = data.Data.PaymentMethod
	o.Details.PaymentID = data.Data.PaymentID

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Omit("Tickets").Save(o).Error; err != nil {
		return err
	}

	return nil
}

func (o *Order) PaymentProcessed(db *gorm.DB, data surfboard_types.OrderWebhookData) error {
	o.Details.PaymentStatus = OrderStatusType(data.Data.PaymentStatus)
	o.Details.PaymentMethod = data.Data.PaymentMethod
	o.Details.PaymentID = data.Data.PaymentID
	o.Details.TruncatedPan = data.Data.TruncatedPan

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Omit("Tickets").Save(o).Error; err != nil {
		return err
	}

	return nil
}

func (o *Order) PaymentCompleted(db *gorm.DB, data surfboard_types.OrderWebhookData) error {
	o.Details.PaymentStatus = OrderStatusType(data.Data.PaymentStatus)
	o.Details.PaymentMethod = data.Data.PaymentMethod
	o.Details.PaymentID = data.Data.PaymentID

	now := time.Now()
	o.Details.PayedAt = &now

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Omit("Tickets").Save(o).Error; err != nil {
		return err
	}

	for _, ticket := range o.Tickets {
		ticket.IsPaid = true
		if err := db.Save(&ticket).Error; err != nil {
			return err
		}
	}

	return nil
}
func (o *Order) PaymentFailed(db *gorm.DB, data surfboard_types.OrderWebhookData) error {
	o.Details.PaymentStatus = OrderStatusType(data.Data.PaymentStatus)
	o.Details.PaymentID = data.Data.PaymentID

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Omit("Tickets").Save(o).Error; err != nil {
		return err
	}

	return nil
}

func (o *Order) PaymentCancelled(db *gorm.DB, data surfboard_types.OrderWebhookData) error {
	o.Details.PaymentStatus = OrderStatusType(data.Data.PaymentStatus)
	o.Details.PaymentID = data.Data.PaymentID

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Omit("Tickets").Save(o).Error; err != nil {
		return err
	}

	return nil
}
