package middleware

import (
	"errors"
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeOrganizationAccess(db *gorm.DB, requiredRole models.OrgRole) gin.HandlerFunc {
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

		authorized, err := CheckUserAuthorization(db, organizationID, c.GetString("user_id"), requiredRole)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check authorization"})
			c.Abort()
			return
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Event not found"})
			c.Abort()
			return
		}

		c.Set("organization", organization) // Store the organization in the context for later use

		c.Next()
	}
}

func CheckUserAuthorization(db *gorm.DB,
	organizationID uint,
	ugkthid string,
	requiredRole models.OrgRole) (bool, error) {
	var err error

	var requestingUser models.User
	var userOrgRole models.OrganizationUserRole

	err = db.Where("id = ?", ugkthid).First(&requestingUser).Error
	if err != nil {
		return false, err
	}

	// Check if the user is a super admin
	if requestingUser.IsSuperAdmin() {
		return true, nil
	}

	err = db.Where("user_ug_kth_id = ? AND organization_id = ?", ugkthid, organizationID).First(&userOrgRole).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User is not found in the organization
			return false, nil
		}
		// Other database error
		return false, err
	}

	var orgRole models.OrganizationRole
	err = db.Where("name  = ?", userOrgRole.OrganizationRoleName).First(&orgRole).Error
	if err != nil {
		return false, err
	}

	var requiredOrgRole models.OrganizationRole
	err = db.Where("name = ?", requiredRole).First(&requiredOrgRole).Error
	if err != nil {
		return false, err
	}

	// Check if the user has the required role
	if orgRole.ID < requiredOrgRole.ID {
		return false, nil
	}

	return true, nil
}
