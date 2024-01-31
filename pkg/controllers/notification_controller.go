package controllers

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationController struct {
	DB *gorm.DB
}

func NewNotificationController(db *gorm.DB) *NotificationController {
	return &NotificationController{DB: db}
}

type EmailRequest struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

func (nc *NotificationController) SendEmail(c *gin.Context) {
	var req EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := nc.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with email not found"})
		return
	}

	err := jobs.HandleTicketAllocationAddToQueue(nc.DB, 1)
	// err = jobs.AddEmailJobToQueue(nc.DB, &user, req.Subject, req.Content, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to add email to queue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}

func (nc *NotificationController) SendTestEmail(c *gin.Context) {
	/*
		Handler that when a ticket allocation is created, it adds a job to the queue
		to send an email to the user that the ticket allocation was created.
	*/
	var user models.User
	if err := nc.DB.Where("email = ?", "lucdow7@gmail.com").First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with email not found"})
		return
	}

	var data = types.EmailTicketAllocationCreated{
		FullName:          "Tsst",
		EventName:         "Blums",
		TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		OrganizationName:  "DKM",
		OrganizationEmail: "test@datasektionen.se",
		PayWithin:         "24",
	}

	tmpl, err := template.ParseFiles("templates/emails/ticket_allocation_created.html")
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}

	// The HTML content of the email is now in buf
	htmlContent := buf.String()

	htmlContent = strings.ReplaceAll(htmlContent, "\x00", "")

	// Create the data to be sent
	jobs.AddEmailJobToQueue(db, &user, "Your ticket to", htmlContent, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}
