package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationController struct {
	notificationService *services.NotificationService
	DB                  *gorm.DB
}

func NewNotificationController(db *gorm.DB, ns *services.NotificationService) *NotificationController {
	return &NotificationController{notificationService: ns, DB: db}
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

	err := jobs.AddEmailJobToQueue(nc.DB, &user, req.Subject, req.Content, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to add email to queue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}
