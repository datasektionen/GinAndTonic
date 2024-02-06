package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
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
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func (cc *ContactController) CreateContact(c *gin.Context) {
	var contact ContactRequest

	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := jobs.SendContactEmail(contact.Name, contact.Email, contact.Subject, contact.Message)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error sending the email"})
		return
	}

	var data types.EmailContact = types.EmailContact{
		FullName: contact.Name,
		Subject:  contact.Subject,
		Message:  contact.Message,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/contact.html", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error sending the email"})
		return
	}

	jobs.SendContactEmail(contact.Name, contact.Email, contact.Subject, htmlContent)

	c.JSON(http.StatusCreated, contact)
}
