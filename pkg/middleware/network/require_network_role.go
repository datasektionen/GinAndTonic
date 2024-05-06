package network_middlewares

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	response_utils "github.com/DowLucas/gin-ticket-release/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireNetworkRole(db *gorm.DB, requiredRole models.NetworkRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(models.User)

		if user.IsSuperAdmin() {
			c.Next()
			return
		}

		networkRole, err := checkUserRole(db, user, requiredRole)
		if err != nil {
			response_utils.RespondWithError(c, 500, err.Error())
			return
		}

		requiredNetworkRole, err := models.GetNetworkRoleByName(db, requiredRole.Name)
		if err != nil {
			response_utils.RespondWithError(c, 500, "Unable to get network role")
			return
		}

		if networkRole.ID < requiredNetworkRole.ID {
			response_utils.RespondWithError(c, 403, "User does not have the required role")
			return
		}

		c.Next()
	}
}

func checkUserRole(db *gorm.DB, user models.User, requiredRole models.NetworkRole) (models.NetworkRole, error) {
	var networkUserRole models.NetworkUserRole

	err := db.Where("user_ug_kth_id = ? AND network_role_name = ?", user.UGKthID, requiredRole.Name).First(&networkUserRole).Error
	if err != nil {
		return models.NetworkRole{}, err
	}

	var role models.NetworkRole

	err = db.Where("name = ?", networkUserRole.NetworkRoleName).First(&role).Error
	if err != nil {
		return models.NetworkRole{}, err
	}

	return role, nil
}
