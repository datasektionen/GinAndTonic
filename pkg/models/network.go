package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Network struct {
	gorm.Model
	Name             string            `json:"name"`
	PlanEnrollmentID uint              `json:"plan_enrollment_id"`
	PlanEnrollment   PlanEnrollment    `json:"plan_enrollment"`
	Users            []User            `gorm:"-" json:"users"`
	NetworkUserRoles []NetworkUserRole `gorm:"foreignKey:NetworkID" json:"network_user_roles"`
	Organizations    []Organization    `gorm:"foreignKey:NetworkID" json:"organizations"`
}

// Get users when the network is loaded
func (n *Network) AfterFind(tx *gorm.DB) (err error) {
	var users []User
	if err := tx.Where("network_id = ?", n.ID).Find(&users).Error; err != nil {
		return err
	}

	n.Users = users
	return nil
}

func GetNetworkByID(db *gorm.DB, id uint) (*Network, error) {
	var network Network
	if err := db.
		Preload("PlanEnrollment.Features.FeatureLimits").
		Preload("PlanEnrollment.FeaturesUsages").
		Preload("NetworkUserRoles").
		Preload("Organizations.Users").
		Preload("Organizations.OrganizationUserRoles").
		First(&network, id).Error; err != nil {
		return &network, err
	}
	return &network, nil
}

func (n Network) AddUserToNetwork(db *gorm.DB, user *User, role NetRole) error {
	// check if user already has a NetworkID
	if user.NetworkID != nil {
		return errors.New("user already belongs to a network")
	}

	// Check if the user is already in the network
	var nur NetworkUserRole
	if err := db.Where("user_ug_kth_id = ? AND network_id = ?", user.UGKthID, n.ID).First(&nur).Error; err == nil {
		return errors.New("user is already in the network")
	}

	nur = NetworkUserRole{
		UserUGKthID:     user.UGKthID,
		NetworkID:       n.ID,
		NetworkRoleName: role,
	}

	if err := db.Create(&nur).Error; err != nil {
		return err
	}

	netID := n.ID
	fmt.Println(netID)
	user.NetworkID = &netID
	if err := db.Save(user).Error; err != nil {
		return err
	}

	return nil
}

func (n Network) RemoveUserFromNetwork(db *gorm.DB, user User) error {
	if err := db.Where("user_ug_kth_id = ? AND network_id = ?", user.UGKthID, n.ID).Delete(&NetworkUserRole{}).Error; err != nil {
		return err
	}
	return nil
}

func (n Network) UserBelongsToNetwork(db *gorm.DB, user User) (bool, error) {
	var nur NetworkUserRole
	if err := db.Where("user_ug_kth_id = ? AND network_id = ?", user.UGKthID, n.ID).First(&nur).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (n Network) CreateNetworkOrganization(db *gorm.DB, organization *Organization, user *User) error {
	organization.NetworkID = n.ID
	if err := db.Create(organization).Error; err != nil {
		return err
	}

	return nil
}
