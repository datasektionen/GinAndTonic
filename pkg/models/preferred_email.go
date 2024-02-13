package models

import (
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

type PreferredEmail struct {
	gorm.Model
	UserUGKthID string     `json:"user_ug_kth_id"`
	Email       string     `json:"email"`
	IsVerified  bool       `json:"is_verified"`
	Token       string     `json:"-"`
	ExpiresAt   *time.Time `json:"-"`
}

// Validate
func ValidateEmail(email string) error {
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	if !match {
		return fmt.Errorf("invalid email format")
	}
	return nil
}
