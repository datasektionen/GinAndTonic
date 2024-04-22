package models

import "gorm.io/gorm"

type FeatureGroupType string
type PaymentPlan string
type PackageTierType string

const (
	FeatureGroupEventManagement     FeatureGroupType = "event_management"
	FeatureGroupTicketManagement    FeatureGroupType = "ticket_management"
	FeatureGroupAPIIntegration      FeatureGroupType = "api_integration"
	FeatureGroupSupport             FeatureGroupType = "support"
	FeatureGroupLandingPage         FeatureGroupType = "landing_page"
	FeatureGroupFinancialManagement FeatureGroupType = "financial_management"
	FeatureGroupEmailManagement     FeatureGroupType = "email_management"
	FeatureGroupOther               FeatureGroupType = "other"
)

const (
	PaymentPlanMonthly PaymentPlan = "monthly"
	PaymentPlanYearly  PaymentPlan = "yearly"
)

const (
	PackageTierFree         PackageTierType = "free"
	PackageTierSingleEvent  PackageTierType = "single_event"
	PackageTierProfessional PackageTierType = "professional"
	PackageTierNetwork      PackageTierType = "network"
)

type PackageTier struct {
	gorm.Model
	Name                 string           `json:"name" gorm:"unique"`
	Tier                 PackageTierType  `json:"tier" gorm:"unique"`
	Description          string           `json:"description"`
	StandardMonthlyPrice int              `json:"standard_monthly_price"` // Monthly amount billed monthly
	StandardYearlyPrice  int              `json:"standard_yearly_price"`  // Monthly amount billed yearly
	PricingPackages      []PricingPackage `gorm:"foreignKey:PackageTierID"`
	DefaultFeatureIDs    []uint           `gorm:"-" json:"default_features"` // Temporary field to hold IDs
	DefaultFeatures      []Feature        `gorm:"many2many:package_tier_default_features;"`
}

type PricingPackage struct {
	gorm.Model
	OrganizationID       *uint       `json:"organization_id"`
	NetworkID            *uint       `json:"network_id"`
	PackageTierID        uint        `json:"package_tier_id" gorm:"not null"`
	Features             []Feature   `gorm:"many2many:package_features;" json:"features"`
	StandardMonthlyPrice int         `json:"monthly_price"` // Monthly amount billed monthly
	StandardYearlyPrice  int         `json:"yearly_price"`  // Monthly amount billed yearly
	Plan                 PaymentPlan `json:"plan" gorm:"not null"`
}

type FeatureGroup struct {
	gorm.Model
	Name        FeatureGroupType `json:"name" gorm:"unique"`
	Description string           `json:"description"`
}

type Feature struct {
	gorm.Model
	Name            string        `json:"name" gorm:"unique"`
	Description     string        `json:"description"`
	FeatureGroupID  uint          `json:"feature_group_id"`
	FeatureGroup    FeatureGroup  `json:"feature_group"`
	IsAvailable     bool          `json:"is_available" gorm:"default:true"` // Indicates if a feature is available in a tier, can be expanded into specifics per tier
	FeatureLimit    FeatureLimit  `gorm:"foreignKey:FeatureID" json:"feature_limit"`
	PackageTiers    []PackageTier `gorm:"many2many:package_tier_default_features;"`
	PackageTiersIDs []uint        `gorm:"-" json:"package_tiers"` // Temporary field to hold IDs
}

type FeatureLimit struct {
	gorm.Model
	FeatureID    uint   `json:"feature_id" gorm:"unique"` // Make this unique
	HardLimit    *int   `json:"hard_limit"`               // Hard limit defines that the limit cannot be exceeded
	MonthlyLimit *int   `json:"monthly_limit"`            // Monthly limit defines the limit per month
	YearlyLimit  *int   `json:"yearly_limit"`             // Yearly limit defines the limit per year
	Description  string `json:"description"`
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
