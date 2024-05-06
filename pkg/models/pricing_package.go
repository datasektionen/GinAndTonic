package models

import (
	"time"

	"gorm.io/gorm"
)

type FeatureGroupType string
type PaymentPlanType string
type PackageTierType string

const (
	FeatureGroupEventManagement     FeatureGroupType = "event_management"
	FeatureGroupTeamManagement      FeatureGroupType = "team_management"
	FeatureGroupTicketManagement    FeatureGroupType = "ticket_management"
	FeatureGroupAPIIntegration      FeatureGroupType = "api_integration"
	FeatureGroupSupport             FeatureGroupType = "support"
	FeatureGroupLandingPage         FeatureGroupType = "landing_page"
	FeatureGroupFinancialManagement FeatureGroupType = "financial_management"
	FeatureGroupEmailManagement     FeatureGroupType = "email_management"
	FeatureGroupOther               FeatureGroupType = "other"
)

const (
	PaymentPlanMonthly PaymentPlanType = "monthly"
	PaymentPlanYearly  PaymentPlanType = "yearly"
	OneTimePayment     PaymentPlanType = "one_time"
	NoPayment          PaymentPlanType = "no_payment"
)

const (
	PackageTierFree         PackageTierType = "free"
	PackageTierSingleEvent  PackageTierType = "single_event"
	PackageTierProfessional PackageTierType = "professional"
	PackageTierNetwork      PackageTierType = "network"
)

type FeatureGroup struct {
	gorm.Model
	Name        FeatureGroupType `json:"name" gorm:"unique"`
	Description string           `json:"description"`
}

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

type PlanEnrollment struct {
	gorm.Model
	ReferenceName  string          `json:"reference_name" gorm:"unique"`
	CreatorEmail   string          `gorm:"-" json:"creator_email"` // Not stored in the database
	CreatorID      string          `json:"creator_id"`
	Creator        User            `json:"creator"`
	PackageTierID  uint            `json:"package_tier_id" gorm:"not null"`
	Features       []Feature       `gorm:"many2many:package_features;" json:"features"`
	MonthlyPrice   int             `json:"monthly_price"`  // Monthly amount billed monthly
	YearlyPrice    int             `json:"yearly_price"`   // Monthly amount billed yearly
	OneTimePrice   int             `json:"one_time_price"` // One time payment
	Plan           PaymentPlanType `json:"plan" gorm:"not null"`
	FeaturesUsages []FeatureUsage  `json:"features_usages" gorm:"foreignKey:PlanEnrollmentID"`
}

type FeatureUsage struct {
	CreatedAt        time.Time `gorm:"primaryKey"`
	FeatureID        uint      `gorm:"primaryKey;autoIncrement:false"`
	PlanEnrollmentID uint      `gorm:"primaryKey;autoIncrement:false"`
	Usage            int       `json:"usage"`
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

type FeatureLimit struct {
	gorm.Model
	FeatureID        uint   `json:"feature_id"`
	PackageTierID    uint   `json:"package_tier_id"`
	LimitDescription string `json:"limit_description"`
	Limit            *int   `json:"limit"` // This is a hard limit
	MonthlyLimit     *int   `json:"monthly_limit"`
	YearlyLimit      *int   `json:"yearly_limit"`
}

func InitializePackageTiers(db *gorm.DB) error {
	// Define the tiers you want to ensure exist
	tiers := []PackageTier{
		{Name: "Free", Tier: PackageTierFree, Description: "A free package for small organizations"},
		{Name: "Single Event", Tier: PackageTierSingleEvent, Description: "A package for organizations hosting a single event"},
		{Name: "Professional", Tier: PackageTierProfessional, Description: "A package for professional organizations"},
		{Name: "Network", Tier: PackageTierNetwork, Description: "A package for bigger organizations with multiple events"},
	}

	// Check each tier and create it if it doesn't exist
	for _, tier := range tiers {
		var existingTier PackageTier
		db.Where("name = ?", tier.Name).FirstOrCreate(&existingTier, tier)
	}
	return nil
}

func InitializeFeatureGroups(db *gorm.DB) error {
	// Define the feature groups you want to ensure exist
	featureGroups := []FeatureGroup{
		{Name: FeatureGroupEventManagement, Description: "Event management features"},
		{Name: FeatureGroupTicketManagement, Description: "Ticket management features"},
		{Name: FeatureGroupTeamManagement, Description: "Team management features"},
		{Name: FeatureGroupAPIIntegration, Description: "API integration features"},
		{Name: FeatureGroupSupport, Description: "Support features"},
		{Name: FeatureGroupLandingPage, Description: "Landing page features"},
		{Name: FeatureGroupFinancialManagement, Description: "Financial management features"},
		{Name: FeatureGroupEmailManagement, Description: "Email management features"},
		{Name: FeatureGroupOther, Description: "Other features"},
	}

	// Check each feature group and create it if it doesn't exist
	for _, featureGroup := range featureGroups {
		var existingFeatureGroup FeatureGroup
		db.Where("name = ?", featureGroup.Name).FirstOrCreate(&existingFeatureGroup, featureGroup)
	}
	return nil
}

func DeleteFeature(db *gorm.DB, id uint) error {
	// Find the feature
	var feature Feature
	if err := db.First(&feature, id).Error; err != nil {
		return err
	}

	// Delete the feature
	if err := db.Delete(&feature).Error; err != nil {
		return err
	}

	return nil
}

func GetPlanEnrollmentByID(db *gorm.DB, id uint) (PlanEnrollment, error) {
	var planEnrollment PlanEnrollment
	err := db.Preload("Features").Preload("FeaturesUsages").Where("id = ?", id).First(&planEnrollment).Error
	return planEnrollment, err
}

func GetPlanEnrollmentByNetworkID(db *gorm.DB, networkID uint) (PlanEnrollment, error) {
	var planEnrollment PlanEnrollment
	err := db.Where("network_id = ?", networkID).First(&planEnrollment).Error
	return planEnrollment, err
}

func GetFeatureByName(db *gorm.DB, name string) (Feature, error) {
	var feature Feature
	err := db.Where("name = ?", name).First(&feature).Error
	return feature, err
}

func GetPackageTier(db *gorm.DB, id uint) (PackageTier, error) {
	var packageTier PackageTier
	err := db.Preload("DefaultFeatures").Where("id = ?", id).First(&packageTier).Error
	return packageTier, err
}

func GetPackageTierByType(db *gorm.DB, tier PackageTierType) (PackageTier, error) {
	var packageTier PackageTier
	err := db.Preload("DefaultFeatures").Where("tier = ?", tier).First(&packageTier).Error
	return packageTier, err
}
