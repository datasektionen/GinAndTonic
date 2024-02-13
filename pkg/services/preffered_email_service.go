package services

import (
	"errors"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

type PreferredEmailService struct {
	DB *gorm.DB
}

func NewPreferredEmailService(db *gorm.DB) *PreferredEmailService {
	return &PreferredEmailService{DB: db}
}

func (pes *PreferredEmailService) RequestPreferredEmailChange(
	user *models.User,
	email string,
) (r *types.ErrorResponse) {
	// Handles a request to change the preffered email
	if !user.VerifiedEmail {
		return &types.ErrorResponse{
			StatusCode: 400,
			Message:    "User email not verified",
		}
	}

	tx := pes.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingUser models.User
	if err := tx.Joins("INNER JOIN preferred_emails ON users.id = preferred_emails.user_id").
		Preload("PrefferedEmail").
		Where("users.ug_kth_id != ? AND (users.email = ? OR preferred_emails.email = ?)", user.UGKthID, email, email).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 500,
				Message:    err.Error(),
			}
		}
	}

	if existingUser.Email == email {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 400,
			Message:    "Email already in use",
		}
	}

	if existingUser.PreferredEmail.IsVerified {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 400,
			Message:    "Email already in use",
		}
	}

	token := utils.GenerateRandomString(32)
	expires := time.Now().Add(time.Hour * 1)

	// If we get this far we can create a new preffered email
	prefferedEmail := models.PreferredEmail{
		UserUGKthID: user.UGKthID,
		Email:       email,
		Token:       token,
		ExpiresAt:   &expires,
		IsVerified:  false,
	}

	if err := tx.Create(&prefferedEmail).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}
	}

	err := tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}
	}

	return nil
}

func (pes *PreferredEmailService) ConfirmPrefferedEmailChange(
	user *models.User,
	token string,
) (r *types.ErrorResponse) {
	// Handles a request to confirm the preffered email change
	tx := pes.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var prefferedEmail models.PreferredEmail
	if err := tx.Where("user_ug_kth_id = ? AND token = ?", user.UGKthID, token).First(&prefferedEmail).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 400,
			Message:    "Invalid token",
		}
	}

	if prefferedEmail.ExpiresAt.Before(time.Now()) {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 400,
			Message:    "Token expired",
		}
	}

	prefferedEmail.IsVerified = true

	if err := tx.Save(&prefferedEmail).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}
	}

	err := tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}
	}

	return nil
}

// type UserUpdateRequest struct {
// 	Email string `json:"email"`
// }

// func (uc *UserController) ChangePrefferedEmail(c *gin.Context) {
// 	// Check if the user is a super admin
// 	ugkthid := c.MustGet("ugkthid").(string)

// 	var req UserUpdateRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if models.ValidateEmail(req.Email) != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
// 		return
// 	}

// 	tx := uc.DB.Begin()

// 	var user models.User
// 	if err := tx.Where("ug_kth_id = ?", ugkthid).First(&user).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if !user.VerifiedEmail {
// 		tx.Rollback()
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "User email not verified"})
// 		return
// 	}

// 	var existingUser models.User
// 	if err := tx.Where("ug_kth_id != ? AND (email = ? OR preffered_email = ?)", ugkthid, req.Email, req.Email).First(&existingUser).Error; err != nil {
// 		if !errors.Is(err, gorm.ErrRecordNotFound) {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 	} else {
// 		tx.Rollback()
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
// 		return
// 	}

// 	if user.IsExternal {
// 		user.Email = req.Email
// 	} else {
// 		user.PrefferedEmail = &req.Email
// 	}

// 	if err := tx.Save(&user).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	tx.Commit()
// 	c.JSON(http.StatusOK, gin.H{"message": "Preffered email updated successfully"})
// }
