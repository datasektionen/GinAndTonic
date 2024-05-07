package feature_middleware

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	response_utils "github.com/DowLucas/gin-ticket-release/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user's package tier
		user := c.MustGet("user").(models.User)

		// Check if the user is an super_admin
		if user.IsSuperAdmin() {
			c.Next()
			return
		}

		if !user.IsEventManager() {
			c.JSON(403, gin.H{"error": "User is not an event manager"})
			c.Abort()
			return
		}

	}
}

// RequireFeatureLimit is a middleware that checks if the user has the required feature
// and if the user has not exceeded the feature limit

func RequireFeatureLimit(db *gorm.DB, featureName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user's package tier
		user := c.MustGet("user").(models.User)
		objRef, exists := c.Get("object_reference")

		var objectReference *string = nil
		if exists {
			objectReference = objRef.(*string)
		}

		// Check if the user is an super_admin
		if user.IsSuperAdmin() {
			c.Next()
			return
		}

		if !user.IsEventManager() {
			response_utils.RespondWithError(c, 403, "User is not an event manager")
			c.Abort()
			return
		}

		planEnrollment := user.Network.PlanEnrollment
		if planEnrollment.ID == 0 {
			response_utils.RespondWithError(c, 500, "User does not have a plan enrollment")
			c.Abort()
			return
		}

		feature, err := models.GetFeature(db, featureName)
		if err != nil {
			response_utils.RespondWithError(c, 500, "Error getting feature")
			c.Abort()
			return
		}

		canUseFeature, err := feature.CanUseLimitedFeature(db, &planEnrollment, objectReference)

		if err != nil {
			response_utils.RespondWithError(c, 500, "Error checking feature usage")
			c.Abort()
			return
		}

		if !canUseFeature {
			response_utils.RespondWithError(c, 403, "Feature limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Set object reference to use in RequireFeatureLimit
func SetParamObjectReference(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		objectReference := c.Param(param)
		if objectReference == "" {
			c.Set("object_reference", nil)
		} else {
			c.Set("object_reference", &objectReference)
		}

		c.Next()
	}
}
