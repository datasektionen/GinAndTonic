package models

import "gorm.io/gorm"

type NetworkSettings struct {
	gorm.Model
	NetworkID   uint   `json:"network_id"`
	MainColor   string `json:"main_color"`
	AccentColor string `json:"accent_color"`
	Logo        string `json:"logo"` // URL to the logo
}
