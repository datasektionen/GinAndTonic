package models

import "gorm.io/gorm"

type FeatureGroup string
type PaymentPlan string
type PackageTierType string

const (
	FeatureGroupEventManagement     FeatureGroup = "event_management"
	FeatureGroupTicketManagement    FeatureGroup = "ticket_management"
	FeatureGroupAPIIntegration      FeatureGroup = "api_integration"
	FeatureGroupSupport             FeatureGroup = "support"
	FeatureGroupLandingPage         FeatureGroup = "landing_page"
	FeatureGroupFinancialManagement FeatureGroup = "financial_management"
	FeatureGroupEmailManagement     FeatureGroup = "email_management"
	FeatureGroupOther               FeatureGroup = "other"
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
}

type PricingPackage struct {
	gorm.Model
	OrganizationID       *uint       `json:"organization_id"`
	NetworkID            *uint       `json:"network_id"`
	PackageTierID        uint        `json:"package_tier_id" gorm:"not null"`
	Features             []Feature   `gorm:"many2many:package_features;"`
	StandardMonthlyPrice int         `json:"monthly_price"` // Monthly amount billed monthly
	StandardYearlyPrice  int         `json:"yearly_price"`  // Monthly amount billed yearly
	Plan                 PaymentPlan `json:"plan" gorm:"not null"`
}

type Feature struct {
	gorm.Model
	Name         string
	Description  string
	Group        FeatureGroup
	IsAvailable  bool         // Indicates if a feature is available in a tier, can be expanded into specifics per tier
	FeatureLimit FeatureLimit `gorm:"foreignKey:FeatureID"`
}

type FeatureLimit struct {
	gorm.Model
	FeatureID    uint   `json:"feature_id"`
	FeatureName  string `json:"feature_name" gorm:"-"`
	MonthlyLimit *int   `json:"monthly_limit"`
	YearlyLimit  *int   `json:"yearly_limit"`
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
