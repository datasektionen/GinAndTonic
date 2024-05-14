package models

import "gorm.io/gorm"

type NetworkDetails struct {
	gorm.Model
	NetworkID  uint   `json:"network_id"`
	Descrition string `json:"description"` // Description of the network
	Language   string `json:"language"`
	Address    string `json:"address"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Email      string `json:"email"` // main email for the network
}
