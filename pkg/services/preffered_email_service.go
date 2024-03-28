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
	// Handles a request to change the preferred email
	if user.IsExternal {
		return &types.ErrorResponse{
			StatusCode: 400,
			Message:    "External users cannot change their preffered email",
		}
	}

	tx := pes.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingUser models.User
	if err := tx.Joins("INNER JOIN preferred_emails ON users.ug_kth_id = preferred_emails.user_ug_kth_id").
		Preload("PreferredEmail").
		Where("users.ug_kth_id = ?", user.UGKthID).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 500,
				Message:    err.Error(),
			}
		}
	}

	if existingUser.UGKthID != "" {
		if existingUser.Email == email {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 400,
				Message:    "Email already in use",
			}
		}
	}

	var existingPrefferedEmail models.PreferredEmail
	if err := tx.Where("user_ug_kth_id = ?", user.UGKthID).First(&existingPrefferedEmail).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 500,
				Message:    err.Error(),
			}
		}
	}

	if existingPrefferedEmail.ID != 0 {
		if existingPrefferedEmail.Email == email && existingPrefferedEmail.IsVerified {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 400,
				Message:    "You already have this email as your preffered email",
			}
		}
	}

	token := utils.GenerateRandomString(32)
	expires := time.Now().Add(time.Hour * 1)

	// If we get this far we can create a new preferred email
	prefferedEmail := models.PreferredEmail{
		UserUGKthID: user.UGKthID,
		Email:       email,
		Token:       token,
		ExpiresAt:   &expires,
		IsVerified:  false,
	}

	if existingPrefferedEmail.ID != 0 {
		// delete the old preferred email
		if err := tx.Delete(&existingPrefferedEmail).Error; err != nil {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 500,
				Message:    err.Error(),
			}
		}

		if err := tx.Create(&prefferedEmail).Error; err != nil {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 500,
				Message:    err.Error(),
			}
		}
	} else {
		if err := tx.Create(&prefferedEmail).Error; err != nil {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 500,
				Message:    err.Error(),
			}
		}
	}

	err := Notify_RequestChangePreferredEmail(tx, user, &prefferedEmail)

	if err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}
	}

	return nil
}

func (pes *PreferredEmailService) ConfirmPrefferedEmailChange(
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
	if err := tx.Where("token = ?", token).First(&prefferedEmail).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 400,
			Message:    "Invalid token",
		}
	}

	if prefferedEmail.IsVerified {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 204,
			Message:    "Email already verified",
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
