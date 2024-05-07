package models

import (
	"gorm.io/gorm"
)

type PlanEnrollment struct {
	gorm.Model
	ReferenceName        string                 `json:"reference_name" gorm:"unique"`
	CreatorEmail         string                 `gorm:"-" json:"creator_email"` // Not stored in the database
	CreatorID            string                 `json:"creator_id" gorm:"foreignkey:UGKthID"`
	Creator              User                   `json:"creator" gorm:"-"`
	RequiredPlanFeatures []RequiredPlanFeatures `json:"required_plan_features" gorm:"-"`
	PackageTierID        uint                   `json:"package_tier_id" gorm:"not null"`
	Features             []*Feature             `gorm:"many2many:package_features;" json:"features"`
	MonthlyPrice         int                    `json:"monthly_price"`  // Monthly amount billed monthly
	YearlyPrice          int                    `json:"yearly_price"`   // Monthly amount billed yearly
	OneTimePrice         int                    `json:"one_time_price"` // One time payment
	Plan                 PaymentPlanType        `json:"plan" gorm:"not null"`
	FeaturesUsages       []FeatureUsage         `json:"features_usages" gorm:"foreignKey:PlanEnrollmentID"`
}

// After finding a value we get the creator
func (pe *PlanEnrollment) AfterFind(tx *gorm.DB) (err error) {
	var creator User
	if err := tx.Where("id = ?", pe.CreatorID).First(&creator).Error; err != nil {
		return err
	}

	pe.Creator = creator

	rpf, err := GetAllRequiredPlanFeatures(tx)
	if err != nil {
		return err
	}

	pe.RequiredPlanFeatures = rpf

	for idx := range pe.Features {
		// Assuming 'feature' is now a *Feature (pointer to Feature)
		f := pe.Features[idx]
		f.HasLimitAccess = true
	}

	return nil
}

// After created
func (pe *PlanEnrollment) AfterCreate(tx *gorm.DB) (err error) {
	var allFeatures []Feature
	if err := tx.Find(&allFeatures).Error; err != nil {
		return err
	}

	for _, feature := range allFeatures {
		var featureUsage FeatureUsage = FeatureUsage{
			FeatureID:        feature.ID,
			PlanEnrollmentID: pe.ID,
			Usage:            0,
		}

		if err := tx.Create(&featureUsage).Error; err != nil {
			return err
		}
	}

	var creator User
	if err := tx.Where("id = ?", pe.CreatorID).First(&creator).Error; err != nil {
		return err
	}

	// Add the role of Manager to the creator to signify that they are manager aswell
	err = creator.AddRole(tx, RoleManager)
	if err != nil {
		return err
	}

	return nil
}

func (pe *PlanEnrollment) ClearFeatures(db *gorm.DB) error {
	if err := db.Model(pe).Association("Features").Clear(); err != nil {
		return err
	}

	return nil
}
