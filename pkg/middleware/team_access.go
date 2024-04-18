package middleware

import (
	"errors"
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeTeamAccess(db *gorm.DB, requiredRole models.OrgRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamIDStr := c.Param("teamID")

		teamID, err := utils.ParseStringToUint(teamIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
			c.Abort()
			return
		}

		var team models.Team
		if err := db.First(&team, teamID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			c.Abort()
			return
		}

		authorized, err := CheckUserAuthorization(db, teamID, c.GetString("ugkthid"), requiredRole)
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

		c.Set("team", team) // Store the team in the context for later use

		c.Next()
	}
}

func CheckUserAuthorization(db *gorm.DB,
	teamID uint,
	ugkthid string,
	requiredRole models.OrgRole) (bool, error) {
	var err error

	var requestingUser models.User
	var userOrgRole models.TeamUserRole

	err = db.Where("ug_kth_id = ?", ugkthid).First(&requestingUser).Error
	if err != nil {
		return false, err
	}

	// Check if the user is a super admin
	if requestingUser.IsSuperAdmin() {
		return true, nil
	}

	err = db.Where("user_ug_kth_id = ? AND team_id = ?", ugkthid, teamID).First(&userOrgRole).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User is not found in the team
			return false, nil
		}
		// Other database error
		return false, err
	}

	var orgRole models.TeamRole
	err = db.Where("name  = ?", userOrgRole.TeamRoleName).First(&orgRole).Error
	if err != nil {
		return false, err
	}

	var requiredOrgRole models.TeamRole
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
