package middleware

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeTeamRole(db *gorm.DB, rr models.OrgRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requiredRole models.TeamRole
		// Obtain the user ID and team ID from the context or request parameters
		UGKthID := c.MustGet("ugkthid").(string)
		teamID, _ := strconv.Atoi(c.Param("teamID"))

		// Check the user's role within the team
		userRole, err := checkUserRole(db, UGKthID, uint(teamID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		requiredRole, err = models.GetTeamRole(db, rr)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Compare the user's role with the required role
		// Owner has the highest id, so if the user's role id is lower than the required role id, the user is not authorized
		if userRole.ID < requiredRole.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "User not authorized for this team"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func checkUserRole(db *gorm.DB, UGKthID string, teamID uint) (models.TeamRole, error) {
	var teamUserRole models.TeamUserRole

	// Assume db is your *gorm.DB object
	err := db.Where("user_ug_kth_id = ? AND team_id = ?", UGKthID, teamID).First(&teamUserRole).Error
	if err != nil {
		return models.TeamRole{}, err
	}

	var role models.TeamRole

	// Get the role
	err = db.Where("name = ?", teamUserRole.TeamRoleName).First(&role).Error
	if err != nil {
		return models.TeamRole{}, err
	}

	return role, nil
}
