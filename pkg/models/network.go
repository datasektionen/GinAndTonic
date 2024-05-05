package models

import "gorm.io/gorm"

type Network struct {
	gorm.Model
	Name             string            `json:"name"`
	PlanEnrollmentID *uint             `json:"plan_enrollment_id"`
	Users            []User            `gorm:"foreignKey:NetworkID" json:"users"`
	NetworkUserRoles []NetworkUserRole `gorm:"foreignKey:NetworkID" json:"network_user_roles"`
	Organizations    []Organization    `gorm:"foreignKey:NetworkID"`
}
