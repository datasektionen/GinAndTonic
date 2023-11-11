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

// AddUserToOrganization handles adding a user to an organization
func (ouc *OrganisationUsersController) AddUserToOrganization(c *gin.Context) {
	userUGKthID, organizationID, err := ouc.parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ouc.OrganisationService.AddUserToOrganization(userUGKthID, organizationID, models.OrganizationMember)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to organization"})
}

// RemoveUserFromOrganization handles removing a user from an organization
func (ouc *OrganisationUsersController) RemoveUserFromOrganization(c *gin.Context) {
	userUGKthID, organizationID, err := ouc.parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ouc.OrganisationService.RemoveUserFromOrganization(userUGKthID, organizationID)
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

// Helper methods

func (ouc *OrganisationUsersController) parseParams(c *gin.Context) (string, uint, error) {
	userUGKthID := c.Param("ugkthid")
	organizationIDStr := c.Param("organizationID")

	organizationID, err := strconv.Atoi(organizationIDStr)
	if err != nil {
		return "", 0, fmt.Errorf("Invalid organization ID")
	}

	return userUGKthID, uint(organizationID), nil
}
