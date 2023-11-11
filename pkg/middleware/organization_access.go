package middleware

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeOrganizationAccess(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		organizationIDStr := c.Param("organizationID")

		organizationID, err := utils.ParseStringToUint(organizationIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			c.Abort()
			return
		}

		var organization models.Organization
		if err := db.First(&organization, organizationID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
			c.Abort()
			return
		}

		authorized, err := CheckUserAuthorization(db, organizationID, c.GetString("ugkthid"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check authorization"})
			c.Abort()
			return
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "User not authorized for this event"})
			c.Abort()
			return
		}

		c.Set("organization", organization) // Store the organization in the context for later use

		c.Next()
	}
}

func CheckUserAuthorization(db *gorm.DB, organizationID uint, ugkthid string) (bool, error) {
	var count int64
	err := db.Table("organization_users").
		Where("organization_id = ? AND user_ug_kth_id = ?", organizationID, ugkthid).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
