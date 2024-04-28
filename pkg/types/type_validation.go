package types

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

// Validates that a field is not empty.
func ValidateNotEmpty(field, fieldName string) *ErrorResponse {
	if field == "" {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    fmt.Sprintf("%s cannot be empty", fieldName),
		}
	}
	return nil
}

func ValidateNameField(field, fieldName string) *ErrorResponse {
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

func CheckEmailNotInUse(db *gorm.DB, email string) *ErrorResponse {
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err == nil {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "email already in use",
		}
	}
	return nil
}

func CheckUGKthIDNotInUse(db *gorm.DB, UGKthID string) *ErrorResponse {
	var user models.User
	if err := db.Where("ug_kth_id = ?", UGKthID).First(&user).Error; err == nil {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "username already in use",
		}
	}
	return nil
}

// Validates the email format.
func ValidateEmail(email string) *ErrorResponse {
	if _, err := mail.ParseAddress(email); err != nil {
		return &ErrorResponse{
			StatusCode: 400,
			Message:    "invalid email address",
		}
	}
	return nil
}

// Validates the password (add your own complexity requirements as needed).
func ValidatePassword(password, passwordRepeat string) *ErrorResponse {
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

func ValidatePhoneNumber(phoneNumber string) *ErrorResponse {
	if len(phoneNumber) < 10 {
		return &ErrorResponse{StatusCode: 400, Message: "phone number must be at least 10 characters long"}
	}
	return nil
}
