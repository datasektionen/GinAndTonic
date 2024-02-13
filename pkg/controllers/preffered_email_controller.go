package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PreferredEmailController struct {
	DB      *gorm.DB
	service *services.PreferredEmailService
}

func NewPreferredEmailController(db *gorm.DB, s *services.PreferredEmailService) *PreferredEmailController {
	return &PreferredEmailController{DB: db, service: s}
}

type PrefferedEmailRequest struct {
	Email string `json:"email" binding:"required"`
}

func (pec *PreferredEmailController) Request(c *gin.Context) {
	// Handles a request to change the preffered email
	var req PrefferedEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ugkthid := c.MustGet("ugkthid").(string)
	var user models.User
	if err := pec.DB.Where("ug_kth_id = ?", ugkthid).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	e := pec.service.RequestPreferredEmailChange(&user, req.Email)

	if e != nil {
		c.JSON(e.StatusCode, gin.H{"error": e.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email change request sent"})
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

func (pec *PreferredEmailController) Verify(c *gin.Context) {
	// Handles a request to confirm the preffered email change
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r := pec.service.ConfirmPrefferedEmailChange(req.Token)

	if r != nil {
		c.JSON(r.StatusCode, gin.H{"error": r.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email change verified"})
}
