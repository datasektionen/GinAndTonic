package models

import "gorm.io/gorm"

type NetworkSettings struct {
	gorm.Model
	NetworkID uint `json:"network_id"`
}
