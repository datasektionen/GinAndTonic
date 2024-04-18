package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrganisationUsersController struct {
	DB                  *gorm.DB
	OrganisationService *services.OrganisationService
}

func NewTeamUsersController(db *gorm.DB, os *services.OrganisationService) *OrganisationUsersController {
	return &OrganisationUsersController{DB: db, OrganisationService: os}
}

// AddUserToTeam handles adding a user to an team
func (ouc *OrganisationUsersController) AddUserToTeam(c *gin.Context) {
	username, teamID, err := ouc.parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ouc.OrganisationService.AddUserToTeam(username, teamID, models.TeamMember)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to team"})
}

// RemoveUserFromTeam handles removing a user from an team
func (ouc *OrganisationUsersController) RemoveUserFromTeam(c *gin.Context) {
	username, teamID, err := ouc.parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ouc.OrganisationService.RemoveUserFromTeam(username, teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from team"})
}

// GetTeamUsers handles fetching users of an team
func (ouc *OrganisationUsersController) GetTeamUsers(c *gin.Context) {
	teamIDStr := c.Param("teamID")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	users, err := ouc.OrganisationService.GetTeamUsers(uint(teamID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (ouc *OrganisationUsersController) ChangeUserTeamRole(c *gin.Context) {
	username, teamID, err := ouc.parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check that the user is not changing their own role
	err = ouc.checkUserNotSelf(c, username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := models.StringToOrgRole(req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ouc.OrganisationService.ChangeUserRoleInTeam(username, teamID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role changed"})
}

// Helper methods

func (ouc *OrganisationUsersController) parseParams(c *gin.Context) (string, uint, error) {
	username := c.Param("username")
	teamIDStr := c.Param("teamID")

	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		return "", 0, fmt.Errorf("Invalid team ID")
	}

	return username, uint(teamID), nil
}

// check that checking user is not the same as the user being checked
func (ouc *OrganisationUsersController) checkUserNotSelf(c *gin.Context, username string) error {
	ugkthid, exists := c.Get("ugkthid")
	if !exists {
		return fmt.Errorf("User not authenticated")
	}

	user, err := models.GetUserByUGKthIDIfExist(ouc.DB, ugkthid.(string))
	if err != nil {
		return err
	}

	if user.Username == username {
		return fmt.Errorf("Cannot change your own role")
	}

	return nil
}
