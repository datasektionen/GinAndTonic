package types

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type ExternalSignupRequest struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
}

type ExternalLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *ExternalSignupRequest) Validate() *ErrorResponse {
	// Validate each field.
	// Print the json object and the error message

	if err := r.ValidateNameField(r.FirstName, "first name"); err != nil {
		return err
	}
	if err := r.ValidateNameField(r.LastName, "last name"); err != nil {
		return err
	}

	if err := validateNotEmpty(r.Email, "email"); err != nil {
		return err
	}
	if err := validateEmail(r.Email); err != nil {
		return err
	}
	if err := validatePassword(r.Password, r.PasswordRepeat); err != nil {
		return err
	}

	return nil
}

func (r *ExternalSignupRequest) ValidateNameField(field, fieldName string) *ErrorResponse {
	if err := validateNotEmpty(field, fieldName); err != nil {
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
func (r *ExternalSignupRequest) CheckEmailNotInUse(db *gorm.DB) *ErrorResponse {
	var user models.User
	if err := db.Where("email = ?", r.Email).First(&user).Error; err == nil {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "email already in use",
		}
	}
	return nil
}

func (r *ExternalSignupRequest) CheckUGKthIDNotInUse(db *gorm.DB, UGKthID string) *ErrorResponse {
	var user models.User
	if err := db.Where("ug_kth_id = ?", UGKthID).First(&user).Error; err == nil {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "username already in use",
		}
	}
	return nil
}

// Validates that a field is not empty.
func validateNotEmpty(field, fieldName string) *ErrorResponse {
	if field == "" {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    fmt.Sprintf("%s cannot be empty", fieldName),
		}
	}
	return nil
}

// Validates the email format.
func validateEmail(email string) *ErrorResponse {
	if _, err := mail.ParseAddress(email); err != nil {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "invalid email address",
		}
	}
	return nil
}

// Validates the username (you can add more rules here).
func validateUsername(username string) *ErrorResponse {
	if err := validateNotEmpty(username, "username"); err != nil {
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

// Validates the password (add your own complexity requirements as needed).
func validatePassword(password, passwordRepeat string) *ErrorResponse {
	if password != passwordRepeat {
		return &ErrorResponse{StatusCode: 400, Message: "passwords do not match"}
	}
	if len(password) < 10 {
		return &ErrorResponse{StatusCode: 400, Message: "password must be at least 10 characters long"}
	}
	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber {
		return &ErrorResponse{StatusCode: 400, Message: "password must contain at least one uppercase letter, one lowercase letter and one number"}
	}
	return nil
}
