package surfboard_controllers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var CERTIFICATE string

func init() {
	CERTIFICATE = os.Getenv("SURFBOARD_WEBHOOK_CERTIFICATE")
}

type PaymentWebhookController struct {
	DB *gorm.DB
}

func NewPaymentWebhookController(db *gorm.DB) *PaymentWebhookController {
	return &PaymentWebhookController{DB: db}
}

type WebhookData struct {
	EventType string `json:"eventType"`
	Metadata  struct {
		EventID      string `json:"eventId"`
		Created      int    `json:"created,omitempty"`
		RetryAttempt int    `json:"retryAttempt"`
	} `json:"metadata"`
	Data struct {
		OrderID                     string `json:"orderId"`
		PaymentID                   string `json:"paymentId,omitempty"`
		OrderStatus                 string `json:"orderStatus,omitempty"`
		TransactionID               string `json:"transactionId,omitempty"`
		PaymentMethod               string `json:"paymentMethod,omitempty"`
		PaymentStatus               string `json:"paymentStatus,omitempty"`
		TruncatedPan                string `json:"truncatedPan,omitempty"`
		CardLabel                   string `json:"cardLabel,omitempty"`
		PosEntryMode                string `json:"posEntryMode,omitempty"`
		IssuerApplication           string `json:"issuerApplication,omitempty"`
		TerminalVerificationResult  string `json:"terminalVerificationResult,omitempty"`
		Aid                         string `json:"aid,omitempty"`
		CustomerResponseCode        string `json:"customerResponseCode,omitempty"`
		CvmMethod                   string `json:"cvmMethod,omitempty"`
		AuthMode                    string `json:"authMode,omitempty"`
		CustomerResponseDescription string `json:"customerResponseDescription,omitempty"`
	} `json:"data,omitempty"`
}

func generateHMACSignature(certificate, message string) string {
	key := []byte(certificate)
	h := hmac.New(sha512.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (pwc *PaymentWebhookController) HandleWebhook(c *gin.Context) {
	var data WebhookData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "OK"})
}
