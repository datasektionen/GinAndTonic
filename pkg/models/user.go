package models

import (
	"time"

	"gorm.io/gorm"
)

// User is a struct that represents a user in the database
type User struct {
	UGKthID     string  `gorm:"column:id;primaryKey;index;not null;unique" json:"id"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phone_number" gorm:"default:NULL"`

	VerifiedEmail           bool       `json:"verified_email"`
	EmailVerificationToken  *string    `gorm:"size:255" json:"-"`
	EmailVerificationSentAt *time.Time `json:"-"`
	PasswordHash            *string    `json:"-" gorm:"column:password_hash;default:NULL"`
	RequestToken            *string    `json:"-" gorm:"column:request_token;default:NULL"` // Used by guest users to make requests

	NetworkID *uint `json:"network_id" gorm:"default:NULL"`
	Network   *Network

	TicketOrders          []TicketOrder          `json:"ticket_orders" gorm:"foreignKey:UserUGKthID"`
	Tickets               []Ticket               `json:"tickets" gorm:"foreignKey:UserUGKthID"`
	Organizations         []Organization         `gorm:"many2many:organization_users;" json:"organizations"`
	OrganizationUserRoles []OrganizationUserRole `gorm:"foreignKey:UserUGKthID" json:"organization_user_roles"`
	FoodPreferences       UserFoodPreference     `gorm:"foreignKey:UserUGKthID" json:"food_preferences"`
	Roles                 []Role                 `gorm:"many2many:user_roles;" json:"roles"`
	NetworkUserRoles      []NetworkUserRole      `gorm:"foreignKey:UserUGKthID" json:"network_user_roles"`

	// Other metrics
	ShowedPostLogin bool `json:"showed_post_login" gorm:"default:false"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
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
func GetUserByUGKthIDIfExist(db *gorm.DB, userId string) (User, error) {
	var user User
	err := db.
		Preload("Roles").
		Preload("Organizations").
		Preload("NetworkUserRoles").
		Preload("OrganizationUserRoles").
		Preload("Network.PlanEnrollment.FeaturesUsages").
		Preload("Network.Merchant").
		Where("id = ?", userId).First(&user).Error
	return user, err
}

// GetUserByEmailIfExists returns a user by email if it exists
func GetUserByEmailIfExists(db *gorm.DB, email string) (User, error) {
	var user User
	err := db.Preload("Roles").Where("email = ?", email).First(&user).Error
	return user, err
}

// FullName returns the full name of the user
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) HasRole(role RoleType) bool {
	for _, r := range u.Roles {
		if r.Name == role {
			return true
		}
	}

	return false
}

func (u *User) IsEventManager() bool {
	for _, r := range u.Roles {
		if r.Name == RoleManager {
			return true
		}
	}

	return false
}

// IsSuperAdmin returns true if the user is a super admin
func (u *User) IsSuperAdmin() bool {
	// Preload role
	for _, r := range u.Roles {
		if r.Name == RoleSuperAdmin {
			return true
		}
	}

	return false
}

func (u *User) IsGuestCustomer(db *gorm.DB) bool {
	for _, r := range u.Roles {
		if r.Name == RoleCustomerGuest {
			return true
		}
	}

	return false
}

func (u *User) IsNetworkSuperAdmin() bool {
	for _, r := range u.NetworkUserRoles {
		if r.NetworkRoleName == NetworkSuperAdmin {
			return true
		}
	}

	return false
}

func (u *User) IsNetworkAdmin() bool {
	for _, r := range u.NetworkUserRoles {
		if r.NetworkRoleName == NetworkAdmin {
			return true
		}
	}

	return false
}

func (u *User) IsNetworkMember() bool {
	for _, r := range u.NetworkUserRoles {
		if r.NetworkRoleName == NetworkMember {
			return true
		}
	}

	return false
}

func GetUserPlanEnrollment(db *gorm.DB, user *User) (plan PlanEnrollment, err error) {
	// One user can only belong to one PlanEnrollment since they can only be manager of one organization or network
	// Since the user cannot belong to different plan enrollments, we can just get the first one

	// Get my network
	if user.NetworkID != nil {
		plan, err = GetPlanEnrollmentByNetworkID(db, *user.NetworkID)
		if err != nil {
			return plan, err
		}
	}

	return plan, nil
}

func (u *User) AddRole(db *gorm.DB, role RoleType) error {
	var r Role
	if err := db.Where("name = ?", role).First(&r).Error; err != nil {
		return err
	}

	err := db.Model(u).Association("Roles").Append(&r)
	return err
}

func (u *User) GetNetwork(db *gorm.DB) (Network, error) {
	var network Network
	err := db.Where("id = ?", u.NetworkID).First(&network).Error
	return network, err
}
