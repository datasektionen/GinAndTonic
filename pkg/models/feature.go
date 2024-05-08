package models

import (
	"fmt"

	"gorm.io/gorm"
)

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
	HasLimitAccess  bool           `gorm:"-" json:"has_limit_access"`
}

// Function that returns type
/*
type RequiredPlanFeatures struct {
	FeatureName string            `json:"feature_name"`
	Plans       []PackageTierType `json:"plans"`
}
*/

func GetFeature(db *gorm.DB, name string) (feature Feature, err error) {
	if err := db.Preload("PackageTiers").Where("name = ?", name).First(&feature).Error; err != nil {
		return feature, err
	}
	return feature, nil
}

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

func (f *Feature) CanUseLimitedFeature(db *gorm.DB, planEnrollment *PlanEnrollment, objectReference *string) (bool, error) {
	var featureLimit FeatureLimit
	if err := db.Where("feature_id = ? AND package_tier_id = ?", f.ID, planEnrollment.PackageTierID).First(&featureLimit).Error; err != nil {
		return false, err
	}

	fmt.Println(f.Name)

	currentUsageModel, err := GetLatestFeatureUsage(db, f.ID, planEnrollment.ID, objectReference)
	if err != nil {
		return false, err
	}

	monthlyUsagesList, err := GetMonthlyFeatureUsages(db, f.ID, planEnrollment.ID, objectReference)
	if err != nil {
		return false, err
	}

	yearlyUsagesList, err := GetYearlyFeatureUsages(db, f.ID, planEnrollment.ID, objectReference)
	if err != nil {
		return false, err
	}

	var currentUsage, monthlyUsage, yearlyUsage int

	if currentUsageModel != nil {
		currentUsage = currentUsageModel.Usage
	}

	if len(monthlyUsagesList) > 0 {
		monthlyUsage = GetTotalFeatureUsage(monthlyUsagesList)
	}

	if len(yearlyUsagesList) > 0 {
		yearlyUsage = GetTotalFeatureUsage(yearlyUsagesList)
	}

	fmt.Println("Current usage: ", currentUsage)
	fmt.Println("Monthly usage: ", monthlyUsage)
	fmt.Println("Yearly usage: ", yearlyUsage)

	/*
		- We now want to check the feature limit against the usages.
	*/

	// Hard limit
	if featureLimit.Limit != nil {
		if currentUsage >= *featureLimit.Limit {
			return false, nil
		}
	}

	// Monthly limit
	if featureLimit.MonthlyLimit != nil && planEnrollment.Plan == PaymentPlanMonthly {
		if monthlyUsage >= *featureLimit.MonthlyLimit {
			return false, nil
		}
	}

	// Yearly limit
	if featureLimit.YearlyLimit != nil && planEnrollment.Plan == PaymentPlanYearly {
		if yearlyUsage >= *featureLimit.YearlyLimit {
			return false, nil
		}
	}

	return true, nil
}
