package models

import "gorm.io/gorm"

type FeatureLimit struct {
	gorm.Model
	FeatureID        uint   `json:"feature_id"`
	PackageTierID    uint   `json:"package_tier_id"`
	LimitDescription string `json:"limit_description"`
	Limit            *int   `json:"limit"` // This is a hard limit
	MonthlyLimit     *int   `json:"monthly_limit"`
	YearlyLimit      *int   `json:"yearly_limit"`
}
