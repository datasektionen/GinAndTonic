package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrganisationController struct {
	DB                  *gorm.DB
	OrganisationService *services.OrganisationService
}

func NewOrganizationController(db *gorm.DB, os *services.OrganisationService) *OrganisationController {
	return &OrganisationController{DB: db, OrganisationService: os}
}

func (ec *OrganisationController) CreateOrganization(c *gin.Context) {
	var organization models.Organization

	if err := c.ShouldBindJSON(&organization); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "There was an error creating the organization"})
		return
	}

	if err := organization.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if organization already exists
	if ec.DB.Where("name = ?", organization.Name).First(&organization).RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization already exists"})
		return
	}

	if err := ec.DB.Create(&organization).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": utils.GetDBError(err)})
		return
	}

	createdByUserUGKthID, exists := c.Get("ugkthid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get username
	var user models.User
	if err := ec.DB.Where("ug_kth_id = ?", createdByUserUGKthID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": utils.GetDBError(err)})
		return
	}

	ec.OrganisationService.AddUserToOrganization(user.Username, organization.ID, models.OrganizationOwner)

	c.JSON(http.StatusCreated, gin.H{"organization": organization})
}

func (ec *OrganisationController) ListOrganizations(c *gin.Context) {
	var organizations []models.Organization

	if err := ec.DB.Find(&organizations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organizations": organizations})
}

func (ec *OrganisationController) ListMyOrganizations(c *gin.Context) {
	ugkthid, exists := c.Get("ugkthid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := models.GetUserByUGKthIDIfExist(ec.DB, ugkthid.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	organizations := user.Organizations

	c.JSON(http.StatusOK, gin.H{"organizations": organizations})
}

func (ec *OrganisationController) GetOrganization(c *gin.Context) {
	var organization models.Organization
	id := c.Param("organizationID")

	if err := ec.DB.Preload("Events").
		First(&organization, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": organization})
}

func (ec *OrganisationController) UpdateOrganization(c *gin.Context) {
	var organization models.Organization
	id := c.Param("organizationID")

	if err := ec.DB.First(&organization, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	if err := c.ShouldBindJSON(&organization); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ec.DB.Save(&organization).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": organization})
}

func (ec *OrganisationController) DeleteOrganization(c *gin.Context) {
	var organization models.Organization
	id := c.Param("organizationID")

	if err := ec.DB.First(&organization, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	if err := ec.DB.Delete(&organization).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": organization})
}

func (oc *OrganisationController) ListOrganizationEvents(c *gin.Context) {
	organizationID, err := strconv.Atoi(c.Param("organizationID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	events, err := models.GetAllOrganizationEvents(oc.DB, uint(organizationID))

	c.JSON(http.StatusOK, gin.H{"events": events})
}
