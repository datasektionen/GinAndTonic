package models

import (
	"fmt"
	"strings"
	"time"

	mrand "math/rand"

	"gorm.io/gorm"
)

// UserPasswordReset is a struct that represents a user password reset in the database
type UserPasswordReset struct {
	gorm.Model
	UserUGKthID string    `json:"user_ugkth_id"`
	User        User      `json:"user"`
	Token       string    `json:"token"`
	Used        bool      `json:"used"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func generateToken() string {
	const characters = "DATSEKIONabcdefghijklmnopqrstuvwxyz0123456789"
	var result strings.Builder
	length := len(characters)

	for i := 0; i < 32; i++ {
		randomIndex := mrand.Intn(length)
		result.WriteByte(characters[randomIndex])
	}

	return result.String()
}

func CreatePasswordReset(db *gorm.DB, user *User) *UserPasswordReset {
	// Create a new password reset
	passwordReset := UserPasswordReset{
		User:      *user,
		Token:     generateToken(),
		Used:      false,
		ExpiresAt: time.Now().Add(time.Minute * 30),
	}

	if err := db.Create(&passwordReset).Error; err != nil {
		fmt.Println(err)
		return nil
	}

	return &passwordReset
}
