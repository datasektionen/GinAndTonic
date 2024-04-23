package types

import (
	"fmt"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type CustomerSignupRequest struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	PhoneNumber    *string `json:"phone_number"`
	IsSaved        bool    `json:"is_saved"` // If the user want to be able to login in the future
	Password       *string `json:"password"`
	PasswordRepeat *string `json:"password_repeat"`
}

type CustomerLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *CustomerSignupRequest) Validate() *ErrorResponse {
	// Validate each field.
	// Print the json object and the error message

	if err := ValidateNameField(r.FirstName, "first name"); err != nil {
		return err
	}
	if err := ValidateNameField(r.LastName, "last name"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(r.Email, "email"); err != nil {
		return err
	}
	if err := ValidateEmail(r.Email); err != nil {
		return err
	}

	if r.Password != nil && *r.Password != "" {
		if err := ValidatePassword(*r.Password, *r.PasswordRepeat); err != nil {
			return err
		}
	}

	return nil
}

func (r *CustomerSignupRequest) ValidateNameField(field, fieldName string) *ErrorResponse {
	if err := ValidateNotEmpty(field, fieldName); err != nil {
		return err
	}
	if len(field) > 50 {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    fmt.Sprintf("%s cannot be longer than 50 characters", fieldName),
		}
	}

	// No special characters except for hyphen and apostrophe
	if strings.ContainsAny(field, "!@#$%^&*()_+={}[]\\|;:'\",.<>?/") {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    fmt.Sprintf("%s cannot contain special characters", fieldName),
		}
	}

	return nil
}
func (r *CustomerSignupRequest) CheckEmailNotInUse(db *gorm.DB) *ErrorResponse {
	var user models.User
	if err := db.Where("email = ?", r.Email).First(&user).Error; err == nil {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "email already in use",
		}
	}
	return nil
}

func (r *CustomerSignupRequest) CheckUGKthIDNotInUse(db *gorm.DB, UGKthID string) *ErrorResponse {
	var user models.User
	if err := db.Where("ug_kth_id = ?", UGKthID).First(&user).Error; err == nil {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "username already in use",
		}
	}
	return nil
}

// Validates the username (you can add more rules here).
func validateUsername(username string) *ErrorResponse {
	if err := ValidateNotEmpty(username, "username"); err != nil {
		return err
	}
	if len(username) < 5 {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "username must be at least 5 characters long",
		}
	}
	if len(username) > 20 {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "username cannot be longer than 20 characters",
		}
	}
	if strings.Contains(username, " ") {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "username cannot contain spaces",
		}
	}
	return nil
}
