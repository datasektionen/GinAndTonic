package middleware

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeOrganizationRole(db *gorm.DB, rr models.OrgRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requiredRole models.OrganizationRole
		// Obtain the user ID and organization ID from the context or request parameters
		UGKthID := c.MustGet("ugkthid").(string)
		organizationID, _ := strconv.Atoi(c.Param("organizationID"))

		// Check the user's role within the organization
		userRole, err := checkUserRole(db, UGKthID, uint(organizationID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		requiredRole, err = models.GetOrganizationRole(db, rr)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Compare the user's role with the required role
		// Owener has the highest id, so if the user's role id is lower than the required role id, the user is not authorized
		if userRole.ID < requiredRole.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "User not authorized for this organization"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func checkUserRole(db *gorm.DB, UGKthID string, organizationID uint) (models.OrganizationRole, error) {
	var organizationUserRole models.OrganizationUserRole

	// Assume db is your *gorm.DB object
	err := db.Where("user_ug_kth_id = ? AND organization_id = ?", UGKthID, organizationID).First(&organizationUserRole).Error
	if err != nil {
		return models.OrganizationRole{}, err
	}

	var role models.OrganizationRole

	// Get the role
	err = db.Where("name = ?", organizationUserRole.OrganizationRoleName).First(&role).Error
	if err != nil {
		return models.OrganizationRole{}, err
	}

	return role, nil
}
