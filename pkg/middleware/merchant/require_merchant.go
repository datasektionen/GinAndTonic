package merchant_middleware

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireMerchant(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(models.User)
		network := user.Network

		var merchant models.NetworkMerchant
		if err := db.Where("network_id = ?", network.ID).First(&merchant).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Merchant not found"})
			c.Abort()
			return
		}

		if !merchant.IsApplicationCompleted() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Merchant application not completed"})
			c.Abort()
			return
		}

		c.Next()
	}
}
