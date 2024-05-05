package feature_middleware

import (
	"errors"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	response_utils "github.com/DowLucas/gin-ticket-release/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MiddlewareRequireFeature(db *gorm.DB, requiredFeatureName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(models.User)

		err := CheckFeature(db, requiredFeatureName, user)
		if err != nil {
			switch err.Error() {
			case "User is not an event manager":
				response_utils.RespondWithError(c, 403, err.Error())
			case "Feature not found":
				response_utils.RespondWithError(c, 404, err.Error())
			case "Unable to get feature", "Unable to get user plan enrollment":
				response_utils.RespondWithError(c, 500, err.Error())
			case "User does not have access to this feature":
				response_utils.RespondWithError(c, 403, err.Error())
			default:
				response_utils.RespondWithError(c, 500, "Unknown error")
			}
			return
		}

		c.Next()
	}
}

func CheckFeature(db *gorm.DB, requiredFeatureName string, user models.User) error {
	if user.IsSuperAdmin() {
		return nil
	}

	if !user.IsEventManager() {
		return errors.New("User is not an event manager")
	}

	if err := checkFeatureExists(db, requiredFeatureName); err != nil {
		return err
	}

	if err := checkUserHasFeature(db, &user, requiredFeatureName); err != nil {
		return err
	}

	return nil
}

func checkFeatureExists(db *gorm.DB, requiredFeatureName string) error {
	_, err := models.GetFeatureByName(db, requiredFeatureName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("Feature not found")
		} else {
			return errors.New("Unable to get feature")
		}
	}
	return nil
}

func checkUserHasFeature(db *gorm.DB, user *models.User, requiredFeatureName string) error {
	planEnrollment, err := models.GetUserPlanEnrollment(db, user)
	if err != nil {
		return errors.New("Unable to get user plan enrollment")
	}

	for _, feature := range planEnrollment.Features {
		if feature.Name == requiredFeatureName {
			return nil
		}
	}

	return errors.New("User does not have access to this feature")
}
