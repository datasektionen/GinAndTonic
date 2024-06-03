package surfboard_webhook_service

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	surfboard_service_receipt "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/receipts"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
	"gorm.io/gorm"
)

type OrderWebhookService struct {
	DB *gorm.DB
}

func NewOrderWebhookService(db *gorm.DB) *OrderWebhookService {
	return &OrderWebhookService{DB: db}
}

func (pws *OrderWebhookService) HandleOrderWebhook(data surfboard_types.OrderWebhookData) *types.ErrorResponse {
	// TODO Validate data?
	var err error
	tx := pws.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	eventType := data.EventType

	fmt.Println("Event type: ", eventType)

	orderID := data.Data.OrderID

	order, err := models.GetOrderByID(pws.DB, orderID)
	if err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}

	if !order.CanUpdateOrder() {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Order cannot be updated",
		}
	}

	switch eventType {
	case surfboard_types.OrderCancelled:
		err = pws.handleOrderCancelled(order)
	case surfboard_types.OrderUpdated:
		break
	case surfboard_types.OrderDeleted:
		break
	case surfboard_types.OrderTerminalEvent:
		break
	case surfboard_types.OrderPaymentInit:
		err = pws.handlePaymentInitiated(order, data)
	case surfboard_types.OrderPaymentProc:
		err = pws.handlePaymentProcessed(order, data)
	case surfboard_types.OrderPaymentComp:
		err = pws.handlePaymentCompleted(order, data)
	case surfboard_types.OrderPaymentFailed:
		err = pws.handlePaymentFailed(order, data)
	case surfboard_types.OrderPaymentCancelled:
		err = pws.handlePaymentCancelled(order, data)
	case surfboard_types.OrderPaymentVoided:
		break
	default:
		return &types.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid event type",
		}
	}

	if err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}

	owe := models.OrderWebhookEvent{
		WebhookEvent: models.WebhookEvent{
			EventType:        string(eventType),
			EventID:          data.Metadata.EventID,
			WebhookCreatedAt: time.Unix(int64(data.Metadata.Created), 0),
			RetryAttempt:     data.Metadata.RetryAttempt, // Assuming this is passed in data
			WebhookEventID:   data.Metadata.WebhookEventId,
		},
		OrderID: orderID,
	}

	// Save the event and process it
	if err := pws.saveWebhookEvent(tx, &owe); err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Error committing transaction",
		}
	}

	return nil
}

func (pws *OrderWebhookService) saveWebhookEvent(tx *gorm.DB, owe *models.OrderWebhookEvent) error {
	// Start transaction
	if err := tx.Create(&owe.WebhookEvent).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Process the specific order webhook logic
	if err := owe.Process(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (pws *OrderWebhookService) handleOrderCancelled(order *models.Order) error {
	err := order.CancelOrder(pws.DB)

	return err
}

func (pws *OrderWebhookService) handlePaymentInitiated(order *models.Order, data surfboard_types.OrderWebhookData) error {
	return order.PaymentInitiated(pws.DB, data)
}

func (pws *OrderWebhookService) handlePaymentProcessed(order *models.Order, data surfboard_types.OrderWebhookData) error {
	return order.PaymentProcessed(pws.DB, data)
}

func (pws *OrderWebhookService) handlePaymentCompleted(order *models.Order, data surfboard_types.OrderWebhookData) error {
	err := order.PaymentCompleted(pws.DB, data)
	if err != nil {
		return err
	}

	err = surfboard_service_receipt.EmailReceipt(pws.DB, order.MerchantID, order.OrderID)

	if err != nil {
		log.Printf("Error sending email receipt: %v", err)
		return nil
	}

	return nil
}

func (pws *OrderWebhookService) handlePaymentFailed(order *models.Order, data surfboard_types.OrderWebhookData) error {
	return order.PaymentFailed(pws.DB, data)
}

func (pws *OrderWebhookService) handlePaymentCancelled(order *models.Order, data surfboard_types.OrderWebhookData) error {
	return order.PaymentCancelled(pws.DB, data)
}
