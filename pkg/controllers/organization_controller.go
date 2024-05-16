package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	feature_services "github.com/DowLucas/gin-ticket-release/pkg/services/features"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrganisationController struct {
	DB                  *gorm.DB
	OrganisationService *services.OrganisationService
}

func NewOrganizationController(db *gorm.DB) *OrganisationController {
	return &OrganisationController{DB: db, OrganisationService: services.NewOrganizationService(db)}
}

func (ec *OrganisationController) CreateNetworkOrganization(c *gin.Context) {
	var organization models.Organization

	user := c.MustGet("user").(models.User)

	network := user.Network

	if network == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not belong to a network"})
		return
	}

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

	// Check if email is in use
	if ec.DB.Where("email = ?", organization.Email).First(&organization).RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
		return
	}

	if err := network.CreateNetworkOrganization(ec.DB, &organization, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ec.OrganisationService.AddUserToOrganization(user.Email, organization.ID, models.OrganizationOwner)

	// Increment the max_teams_per_network
	networkID := fmt.Sprintf("%d", *user.NetworkID)                                                                           // Convert req.EventID to string
	err := feature_services.IncrementFeatureUsage(ec.DB, user.Network.PlanEnrollment.ID, "max_teams_per_network", &networkID) // Pass the address of eventID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	ugkthid, exists := c.Get("user_id")
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

	for _, organization := range organizations {
		fmt.Println(organization.CommonEventLocations)
	}

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

type UpdateOrganizationRequest struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (ec *OrganisationController) UpdateOrganization(c *gin.Context) {
	var req UpdateOrganizationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var organization models.Organization
	id := c.Param("organizationID")

	ID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	if err := ec.DB.First(&organization, ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	if req.Email != organization.Email {
		var existingOrganization models.Organization
		if ec.DB.Where("email = ?  ", req.Email).First(&existingOrganization).RowsAffected > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
			return
		}
	}

	if req.Name != organization.Name {
		var existingOrganization models.Organization
		if ec.DB.Where("name = ?  ", req.Name).First(&existingOrganization).RowsAffected > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name already in use"})
			return
		}
	}

	organization.Email = req.Email
	organization.Name = req.Name

	if err := ec.DB.Save(&organization).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": organization})
}

func (ec *OrganisationController) DeleteOrganization(c *gin.Context) {
	var organization models.Organization
	id := c.Param("organizationID")
	user := c.MustGet("user").(models.User)

	if err := ec.DB.First(&organization, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	if err := ec.DB.Delete(&organization).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	networkID := fmt.Sprintf("%d", *user.NetworkID)                                                                           // Convert req.EventID to string
	err := feature_services.DecrementFeatureUsage(ec.DB, user.Network.PlanEnrollment.ID, "max_teams_per_network", &networkID) // Pass the address of eventID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
