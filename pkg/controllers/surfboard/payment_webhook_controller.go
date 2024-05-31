package surfboard_controllers

import (
	"net/http"
	"os"

	surfboard_webhook_service "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/webhook"
	surfboard_types "github.com/DowLucas/gin-ticket-release/pkg/types/surfboard"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var CERTIFICATE string

func init() {
	CERTIFICATE = os.Getenv("SURFBOARD_WEBHOOK_CERTIFICATE")
}

type PaymentWebhookController struct {
	DB      *gorm.DB
	service *surfboard_webhook_service.OrderWebhookService
}

func NewPaymentWebhookController(db *gorm.DB) *PaymentWebhookController {
	return &PaymentWebhookController{DB: db,
		service: surfboard_webhook_service.NewOrderWebhookService(db)}
}

func (pwc *PaymentWebhookController) HandlePaymentWebhook(c *gin.Context) {
	var data surfboard_types.OrderWebhookData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := pwc.service.HandleOrderWebhook(data)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Webhook processed successfully"})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Webhook processed successfully",
	})
}
