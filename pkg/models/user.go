package models

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// User is a struct that represents a user in the database
type User struct {
	UGKthID        string          `gorm:"primaryKey" json:"ug_kth_id"`
	Username       string          `gorm:"uniqueIndex" json:"username"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Email          string          `gorm:"uniqueIndex" json:"email"`
	PreferredEmail *PreferredEmail `gorm:"foreignKey:UserUGKthID" json:"preferred_email"`

	IsExternal              bool       `gorm:"default:false" json:"is_external"` // External users do not have a KTH account
	VerifiedEmail           bool       `json:"verified_email"`
	EmailVerificationToken  string     `gorm:"size:255" json:"-"`
	EmailVerificationSentAt *time.Time `json:"-"`
	PasswordHash            *string    `json:"-" gorm:"column:password_hash;default:NULL" json:"-"`

	Tickets               []Ticket               `json:"tickets"`
	TicketRequests        []TicketRequest        `gorm:"foreignKey:UserUGKthID" json:"ticket_requests"`
	Organizations         []Organization         `gorm:"many2many:organization_users;" json:"organizations"`
	OrganizationUserRoles []OrganizationUserRole `gorm:"foreignKey:UserUGKthID" json:"organization_user_roles"`
	FoodPreferences       UserFoodPreference     `gorm:"foreignKey:UserUGKthID" json:"food_preferences"`
	RoleID                uint                   `json:"role_id"`
	Role                  Role                   `json:"role"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	DeletedAt             gorm.DeletedAt         `gorm:"index" json:"deleted_at"`
}

func CreateUserIfNotExist(db *gorm.DB, user User) error {
	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Create user
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 2. Create associated user food preference
	userFoodPreference := UserFoodPreference{
		UserUGKthID: user.UGKthID,
	}
	if err := tx.Create(&userFoodPreference).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	tx.Commit()

	return nil
}

// GetUserByUGKthIDIfExist returns a user by UGKthID if it exists
func GetUserByUGKthIDIfExist(db *gorm.DB, UGKthID string) (User, error) {
	var user User
	err := db.
		Preload("Role").
		Preload("Organizations").
		Preload("PreferredEmail").
		Where("ug_kth_id = ?", UGKthID).First(&user).Error
	return user, err
}

// GetUserByEmailIfExists returns a user by email if it exists
func GetUserByEmailIfExists(db *gorm.DB, email string) (User, error) {
	var user User
	err := db.Preload("Role").Where("email = ?", email).First(&user).Error
	return user, err
}
func (u *User) GetUserEmail(db *gorm.DB) string {
	// Get preferred email if it exists and is verified
	var pe PreferredEmail
	if err := db.Where("user_ug_kth_id = ?", u.UGKthID).First(&pe).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Preferred email not found, return user's default email
			return u.Email
		}
		// Log other database errors
		fmt.Println(err.Error())
		return u.Email
	}

	if !pe.IsVerified {
		// Preferred email is not verified, return user's default email
		return u.Email
	}

	// Preferred email is found and verified, return it
	return pe.Email
}

// FullName returns the full name of the user
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsSuperAdmin returns true if the user is a super admin
func (u *User) IsSuperAdmin() bool {
	// Preload role
	return u.RoleID == 1
}
