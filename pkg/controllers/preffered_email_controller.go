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
	if err := pec.DB.Where("ugkthid = ?", ugkthid).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r := pec.service.RequestPreferredEmailChange(&user, req.Email)

	if r != nil {
		c.JSON(r.StatusCode, gin.H{"error": r.Message})
		return
	}

}
