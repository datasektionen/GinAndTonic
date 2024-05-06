package models

import (
	"errors"

	"gorm.io/gorm"
)

type Network struct {
	gorm.Model
	Name             string            `json:"name"`
	PlanEnrollmentID uint              `json:"plan_enrollment_id"`
	PlanEnrollment   PlanEnrollment    `json:"plan_enrollment"`
	Users            []User            `gorm:"foreignKey:NetworkID" json:"users"`
	NetworkUserRoles []NetworkUserRole `gorm:"foreignKey:NetworkID" json:"network_user_roles"`
	Organizations    []Organization    `gorm:"foreignKey:NetworkID" json:"organizations"`
}

func GetNetworkByID(db *gorm.DB, id uint) (*Network, error) {
	var network Network
	if err := db.
		Preload("PlanEnrollment.Features").
		Preload("PlanEnrollment.FeaturesUsages").
		Preload("Users").
		Preload("NetworkUserRoles").
		Preload("Organizations.Users").
		Preload("Organizations.OrganizationUserRoles").
		First(&network, id).Error; err != nil {
		return &network, err
	}
	return &network, nil
}

func (n Network) AddUserToNetwork(db *gorm.DB, user User, role NetRole) error {
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

	user.NetworkID = &n.ID
	if err := db.Save(&user).Error; err != nil {
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
