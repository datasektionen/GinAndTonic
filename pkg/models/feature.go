package models

import "gorm.io/gorm"

// List of features and the plan they belong
// Not a table in the database
type RequiredPlanFeatures struct {
	FeatureName string            `json:"feature_name"`
	Plans       []PackageTierType `json:"plans"`
}

type Feature struct {
	gorm.Model
	Name            string         `json:"name" gorm:"unique"`
	Description     string         `json:"description"`
	FeatureGroupID  uint           `json:"feature_group_id"`
	FeatureGroup    FeatureGroup   `json:"feature_group"`
	IsAvailable     bool           `json:"is_available" gorm:"default:true"` // Indicates if a feature is available in a tier, can be expanded into specifics per tier
	PackageTiers    []PackageTier  `gorm:"many2many:package_tier_default_features;"`
	PackageTiersIDs []uint         `gorm:"-" json:"package_tiers"` // Temporary field to hold IDs
	FeatureLimits   []FeatureLimit `json:"feature_limits" gorm:"foreignKey:FeatureID"`
}

// Function that returns type
/*
type RequiredPlanFeatures struct {
	FeatureName string            `json:"feature_name"`
	Plans       []PackageTierType `json:"plans"`
}
*/

func GetAllRequiredPlanFeatures(db *gorm.DB) (requiredPlanFeatures []RequiredPlanFeatures, err error) {
	var features []Feature
	if err := db.Find(&features).Error; err != nil {
		return nil, err
	}

	for _, feature := range features {
		var packageTiers []PackageTier
		if err := db.Model(&feature).Association("PackageTiers").Find(&packageTiers); err != nil {
			return nil, err
		}

		var plans []PackageTierType
		for _, packageTier := range packageTiers {
			plans = append(plans, packageTier.Tier)
		}

		requiredPlanFeatures = append(requiredPlanFeatures, RequiredPlanFeatures{
			FeatureName: feature.Name,
			Plans:       plans,
		})
	}

	return requiredPlanFeatures, nil
}
