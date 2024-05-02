package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContactController struct {
	DB *gorm.DB
}

// NewContactcontroller creates a new controller with the given database client
func NewContactController(db *gorm.DB) *ContactController {
	return &ContactController{
		DB: db,
	}
}

type ContactRequest struct {
	Name           string `json:"name" binding:"required"`
	OrganizationID int    `json:"organization_id" binding:"required"`
	Email          string `json:"email" binding:"required"`
	Subject        string `json:"subject" binding:"required"`
	Message        string `json:"message" binding:"required"`
}

func (cc *ContactController) CreateContact(c *gin.Context) {
	var contact ContactRequest

	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get organization
	var organization models.Organization
	if err := cc.DB.First(&organization, contact.OrganizationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	var data types.EmailContact = types.EmailContact{
		FullName:         contact.Name,
		OrganizationName: organization.Name,
		Subject:          contact.Subject,
		Message:          contact.Message,
		Email:            contact.Email,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/contact.html", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error sending the email"})
		return
	}

	jobs.SendContactEmail(contact.Name, organization.Email, contact.Email, contact.Subject, htmlContent)

	c.JSON(http.StatusOK, contact)
}

type PlanContactRequest struct {
	Name    string                 `json:"name" binding:"required"`
	Email   string                 `json:"email" binding:"required"`
	Plan    models.PackageTierType `json:"plan" binding:"required"`
	Message string                 `json:"message" binding:"required"`
}

func (cc *ContactController) CreatePlanContact(c *gin.Context) {
	var contact PlanContactRequest

	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var data types.EmailPlanContact = types.EmailPlanContact{
		FullName: contact.Name,
		Plan:     contact.Plan,
		Message:  contact.Message,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/plan_enrollment_contact.html", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error sending the email"})
		return
	}

	to := "lucdow7@gmail.com"

	jobs.SendPlanContactEmail(contact.Plan, contact.Name, to, contact.Email, htmlContent)

	c.JSON(http.StatusOK, contact)
}
