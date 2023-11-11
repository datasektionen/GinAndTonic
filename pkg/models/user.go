package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	UGKthID               string                 `gorm:"primaryKey" json:"ug_kth_id"`
	Username              string                 `gorm:"uniqueIndex" json:"username"`
	FirstName             string                 `json:"first_name"`
	LastName              string                 `json:"last_name"`
	Email                 string                 `gorm:"uniqueIndex" json:"email"`
	Tickets               []Ticket               `json:"tickets"`
	TicketRequests        []TicketRequest        `gorm:"foreignKey:UserUGKthID" json:"ticket_requests"`
	Organizations         []Organization         `gorm:"many2many:organization_users;" json:"organizations"`
	OrganizationUserRoles []OrganizationUserRole `gorm:"foreignKey:UserUGKthID" json:"organization_user_roles"`
	RoleID                uint                   `json:"role_id"`
	Role                  Role                   `json:"role"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	DeletedAt             gorm.DeletedAt         `gorm:"index" json:"deleted_at"`
}

func CreateUserIfNotExist(db *gorm.DB, user User) error {
	// Start transaction
	tx := db.Begin()
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

func GetUserByUGKthIDIfExist(db *gorm.DB, UGKthID string) (User, error) {
	var user User
	err := db.Preload("Role").Where("ug_kth_id = ?", UGKthID).First(&user).Error
	return user, err
}

func GetUserByEmailIfExists(db *gorm.DB, email string) (User, error) {
	var user User
	err := db.Preload("Role").Where("email = ?", email).First(&user).Error
	return user, err
}
