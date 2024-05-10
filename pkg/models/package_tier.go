package models

import "gorm.io/gorm"

type PackageTier struct {
	gorm.Model
	Name                 string           `json:"name" gorm:"unique"`
	Tier                 PackageTierType  `json:"tier" gorm:"unique"`
	Description          string           `json:"description"`
	StandardMonthlyPrice int              `json:"standard_monthly_price"` // Monthly amount billed monthly
	StandardYearlyPrice  int              `json:"standard_yearly_price"`  // Monthly amount billed yearly
	PlanEnrollments      []PlanEnrollment `gorm:"foreignKey:PackageTierID" json:"plan_enrollments"`
	DefaultFeatureIDs    []uint           `gorm:"-" json:"default_features"` // Temporary field to hold IDs
	DefaultFeatures      []Feature        `gorm:"many2many:package_tier_default_features;"`
}

func (pt *PackageTier) GetDefaultFeatures(tx *gorm.DB) (defaultFeatures []*Feature, err error) {
	if err := tx.Model(pt).Association("DefaultFeatures").Find(&defaultFeatures); err != nil {
		return nil, err
	}

	return defaultFeatures, nil
}
