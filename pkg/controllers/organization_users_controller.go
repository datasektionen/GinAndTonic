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

func NewOrganizationUsersController(db *gorm.DB, os *services.OrganisationService) *OrganisationUsersController {
	return &OrganisationUsersController{DB: db, OrganisationService: os}
}

type AddUserToOrganizationRequest struct {
	Email string `json:"email" binding:"required"`
}

// AddUserToOrganization handles adding a user to an organization
func (ouc *OrganisationUsersController) AddUserToOrganization(c *gin.Context) {
	organizationID, err := ouc.parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req AddUserToOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ouc.OrganisationService.AddUserToOrganization(req.Email, organizationID, models.OrganizationMember)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to organization"})
}

// RemoveUserFromOrganization handles removing a user from an organization
func (ouc *OrganisationUsersController) RemoveUserFromOrganization(c *gin.Context) {
	organizationID, err := ouc.parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req AddUserToOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ouc.OrganisationService.RemoveUserFromOrganization(req.Email, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from organization"})
}

// GetOrganizationUsers handles fetching users of an organization
func (ouc *OrganisationUsersController) GetOrganizationUsers(c *gin.Context) {
	organizationIDStr := c.Param("organizationID")
	organizationID, err := strconv.Atoi(organizationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	users, err := ouc.OrganisationService.GetOrganizationUsers(uint(organizationID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (ouc *OrganisationUsersController) ChangeUserOrganizationRole(c *gin.Context) {
	organizationID, err := ouc.parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required"`
		Role  string `json:"role" binding:"required"`
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

	// Check that the user is not changing their own role
	err = ouc.checkUserNotSelf(c, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ouc.OrganisationService.ChangeUserRoleInOrganization(req.Email, organizationID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role changed"})
}

// Helper methods

func (ouc *OrganisationUsersController) parseParams(c *gin.Context) (uint, error) {
	organizationIDStr := c.Param("organizationID")

	if organizationIDStr == "" {
		return 0, fmt.Errorf("organization ID not provided")
	}

	organizationID, err := strconv.Atoi(organizationIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid organization ID")
	}

	return uint(organizationID), nil
}

// check that checking user is not the same as the user being checked
func (ouc *OrganisationUsersController) checkUserNotSelf(c *gin.Context, email string) error {
	ugkthid, exists := c.Get("user_id")
	if !exists {
		return fmt.Errorf("user not authenticated")
	}

	user, err := models.GetUserByUGKthIDIfExist(ouc.DB, ugkthid.(string))
	if err != nil {
		return err
	}

	if user.Email == email {
		return fmt.Errorf("cannot change your own role")
	}

	return nil
}
