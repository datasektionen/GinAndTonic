package models

import "gorm.io/gorm"

type Network struct {
	gorm.Model
	Name          string         `json:"name"`
	Organizations []Organization `gorm:"foreignKey:NetworkID"`
}
