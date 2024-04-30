package services

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

type CustomerAuthService struct {
	DB *gorm.DB
}

func NewCustomerAuthService(db *gorm.DB) *CustomerAuthService {
	return &CustomerAuthService{DB: db}
}

func (eas *CustomerAuthService) ValidateSignupRequest(esr types.CustomerSignupRequest) *types.ErrorResponse {
	// Check passwords match
	if err := esr.Validate(); err != nil {
		return err
	}

	// Check email is not already in use
	if err := esr.CheckEmailNotInUse(eas.DB); err != nil {
		return err
	}

	newUGKthID := fmt.Sprintf("user-%s", utils.GenerateRandomString(8))

	// Check UGKthID is not already in use
	if err := esr.CheckUGKthIDNotInUse(eas.DB, newUGKthID); err != nil {
		return err
	}

	// Get "customer" role
	var role models.Role
	if err := eas.DB.Where("name = ?", models.RoleCustomer).First(&role).Error; err != nil {
		fmt.Println(err)
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error",
		}
	}

	return nil
}
