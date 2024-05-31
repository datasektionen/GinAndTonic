package surfboard_middlware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var CERTIFICATE string

func init() {
	CERTIFICATE = os.Getenv("SURFBOARD_WEBHOOK_CERTIFICATE")
}

func generateHMACSignature(certificate, message string) string {
	key := []byte(certificate)
	h := hmac.New(sha512.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func ValidatePaymentWebhookSignature() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the body
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading request body"})
			c.Abort()
			return
		}

		// Store body for later
		c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonData))

		// Compute HMAC
		bodyStr := string(jsonData)
		expectedSignature := generateHMACSignature(CERTIFICATE, bodyStr)

		// Check the signature from the header
		signature := c.GetHeader("X-Webhook-Signature")
		if signature != expectedSignature {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			c.Abort()
			return
		}

		// Allow request to proceed
		c.Next()
	}
}
