package controllers

import (
	"net/http"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserPasswordResetController is the controller for handling password resets
type UserPasswordResetController struct {
	DB *gorm.DB
}

// NewUserPasswordResetController creates a new controller with the given database client
func NewUserPasswordResetController(db *gorm.DB) *UserPasswordResetController {
	return &UserPasswordResetController{
		DB: db,
	}
}

// UserPasswordResetRequest is the request for creating a password reset
type UserPasswordResetRequest struct {
	Email string `json:"email" binding:"required"`
}

// CreatePasswordReset handles the creation of a password reset
func (uprc *UserPasswordResetController) CreatePasswordReset(c *gin.Context) {
	var user models.User
	var req UserPasswordResetRequest
	_, exists := c.Get("ugkthid")

	if exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Password reset not allowed for authenticated users"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := uprc.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !user.VerifiedEmail {
		c.JSON(http.StatusForbidden, gin.H{"error": "Password reset not allowed for unverified email"})
		return
	}

	if !user.IsExternal {
		c.JSON(http.StatusForbidden, gin.H{"error": "Password reset not allowed for internal users"})
		return
	}

	passwordReset := models.CreatePasswordReset(uprc.DB, &user)
	if passwordReset == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error creating the password reset"})
		return
	}

	err := services.Notify_PasswordReset(uprc.DB, passwordReset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error sending the email"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Password reset created"})
}

// UserPasswordResetCompleteRequest is the request for completing a password reset
type UserPasswordResetCompleteRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// CompletePasswordReset handles the completion of a password reset
func (uprc *UserPasswordResetController) CompletePasswordReset(c *gin.Context) {
	var req UserPasswordResetCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordReset := models.UserPasswordReset{}
	if err := uprc.DB.Preload("User").Where("token = ?", req.Token).First(&passwordReset).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Password reset not found"})
		return
	}

	// check that the password reset has not been used and is within ExpiresAt
	if passwordReset.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Password reset expired"})
		return
	}

	if passwordReset.Used {
		c.JSON(http.StatusForbidden, gin.H{"error": "Password reset already used"})
		return
	}

	user := passwordReset.User

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error hashing the password"})
		return
	}

	user.PasswordHash = &hash

	if err := uprc.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error saving the user"})
		return
	}

	passwordReset.Used = true
	if err := uprc.DB.Save(&passwordReset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error saving the password reset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset completed"})
}
