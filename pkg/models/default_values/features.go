package model_default_values

import (
	"log"
	"os"

	m "github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func GetFeatureGroupID(db *gorm.DB, name m.FeatureGroupType) uint {
	var featureGroup m.FeatureGroup
	db.Where("name = ?", name).First(&featureGroup)
	return featureGroup.ID
}

func GetPackageTierIDs(db *gorm.DB, names []m.PackageTierType) []uint {
	var packageTiers []m.PackageTier
	var packageTierIDs []uint

	db.Where("tier IN (?)", names).Find(&packageTiers)

	for _, packageTier := range packageTiers {
		packageTierIDs = append(packageTierIDs, packageTier.ID)
	}

	return packageTierIDs
}

func GetPackageTierID(db *gorm.DB, name m.PackageTierType) uint {
	var packageTier m.PackageTier
	db.Where("tier = ?", name).First(&packageTier)
	return packageTier.ID
}

func fetchIDsForGroupsAndTiers(db *gorm.DB) (map[m.FeatureGroupType]uint, map[m.PackageTierType]uint) {
	// Fetch all groups and tiers once and store their IDs in maps
	var featureGroups []m.FeatureGroup
	var packageTiers []m.PackageTier
	groupIDs := make(map[m.FeatureGroupType]uint)
	tierIDs := make(map[m.PackageTierType]uint)

	db.Find(&featureGroups)
	db.Find(&packageTiers)

	for _, group := range featureGroups {
		groupIDs[group.Name] = group.ID
	}

	for _, tier := range packageTiers {
		tierIDs[tier.Tier] = tier.ID
	}

	return groupIDs, tierIDs
}

func pID(id int) *int {
	return &id
}

func DefaultFeatures(db *gorm.DB) []m.Feature {
	groupIDs, tierIDs := fetchIDsForGroupsAndTiers(db)
	allIds := []uint{tierIDs[m.PackageTierFree], tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]}
	allIdsExceptFree := []uint{tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]}

	features := []m.Feature{
		{
			Name:            "support",
			Description:     "Support for the application",
			FeatureGroupID:  groupIDs[m.FeatureGroupSupport],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierFree], tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree], LimitDescription: "Limited"},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent], LimitDescription: "Limited"},
				{PackageTierID: tierIDs[m.PackageTierProfessional], LimitDescription: "Premium"},
				{PackageTierID: tierIDs[m.PackageTierNetwork], LimitDescription: "Dedicated"},
			},
		},
		{
			Name:            "email_credits",
			Description:     "Email credits for sending emails",
			FeatureGroupID:  groupIDs[m.FeatureGroupEmailManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierFree], tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree], MonthlyLimit: pID(200), YearlyLimit: pID(200)},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent], MonthlyLimit: pID(500), YearlyLimit: pID(500)},
				{PackageTierID: tierIDs[m.PackageTierProfessional], MonthlyLimit: pID(2500), YearlyLimit: pID(30000)},
				{PackageTierID: tierIDs[m.PackageTierNetwork], LimitDescription: "Custom"},
			},
		},
		{
			Name:            "max_teams_per_network",
			Description:     "Maximum number of teams per network",
			FeatureGroupID:  groupIDs[m.FeatureGroupTeamManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierFree], tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree], Limit: pID(1), LimitDescription: "1"},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent], Limit: pID(1), LimitDescription: "1"},
				{PackageTierID: tierIDs[m.PackageTierProfessional], Limit: pID(5), LimitDescription: "5"},
				{PackageTierID: tierIDs[m.PackageTierNetwork], LimitDescription: "Custom"},
			},
		},
		{
			Name:            "max_events",
			Description:     "Maximum number of events",
			FeatureGroupID:  groupIDs[m.FeatureGroupEventManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierFree], tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree], Limit: pID(1), LimitDescription: "1"},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent], Limit: pID(1), LimitDescription: "1"},
				{PackageTierID: tierIDs[m.PackageTierProfessional], LimitDescription: "Unlimited"},
				{PackageTierID: tierIDs[m.PackageTierNetwork], LimitDescription: "Unlimited"},
			},
		},
		{
			Name:            "max_ticket_addons_per_ticket",
			Description:     "Maximum number of ticket addons per ticket",
			FeatureGroupID:  groupIDs[m.FeatureGroupTicketManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierFree], tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree], Limit: pID(0), LimitDescription: "0"},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent], Limit: pID(3), LimitDescription: "3"},
				{PackageTierID: tierIDs[m.PackageTierProfessional], LimitDescription: "Unlimited"},
				{PackageTierID: tierIDs[m.PackageTierNetwork], LimitDescription: "Custom"},
			},
		},
		{
			Name:            "simple_event_forms",
			Description:     "Simple event forms",
			FeatureGroupID:  groupIDs[m.FeatureGroupEventManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierFree], tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree]},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "advanced_event_forms",
			Description:     "Advanced event forms",
			FeatureGroupID:  groupIDs[m.FeatureGroupEventManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "check_in",
			Description:     "Check-in for events",
			FeatureGroupID:  groupIDs[m.FeatureGroupEventManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "sales_reports",
			Description:     "Sales reports for events",
			FeatureGroupID:  groupIDs[m.FeatureGroupFinancialManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "send_outs",
			Description:     "Send custom emails to attendees (Uses email credits)",
			FeatureGroupID:  groupIDs[m.FeatureGroupEmailManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierSingleEvent], tierIDs[m.PackageTierProfessional], tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "basic_ticket_release_methods",
			Description:     "Basic ticket release methods",
			FeatureGroupID:  groupIDs[m.FeatureGroupTicketManagement],
			IsAvailable:     true,
			PackageTiersIDs: allIds,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree]},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "advanced_ticket_release_methods",
			Description:     "Advanced ticket release methods",
			FeatureGroupID:  groupIDs[m.FeatureGroupTicketManagement],
			IsAvailable:     true,
			PackageTiersIDs: allIdsExceptFree,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "custom_ticket_release_methods",
			Description:     "Create custom ticket release methods",
			FeatureGroupID:  groupIDs[m.FeatureGroupTicketManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "private_events",
			Description:     "Create private events",
			FeatureGroupID:  groupIDs[m.FeatureGroupEventManagement],
			IsAvailable:     true,
			PackageTiersIDs: allIdsExceptFree,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "api_integration",
			Description:     "API integration",
			FeatureGroupID:  groupIDs[m.FeatureGroupAPIIntegration],
			IsAvailable:     false,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "contact_database",
			Description:     "Contact database for storing attendee information",
			FeatureGroupID:  groupIDs[m.FeatureGroupEventManagement],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "custom_domains",
			Description:     "Custom domains for your events",
			FeatureGroupID:  groupIDs[m.FeatureGroupOther],
			IsAvailable:     true,
			PackageTiersIDs: allIdsExceptFree,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "basic_event_site",
			Description:     "Use tessera's integrated event site",
			FeatureGroupID:  groupIDs[m.FeatureGroupLandingPage],
			IsAvailable:     true,
			PackageTiersIDs: allIds,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree]},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "custom_event_site",
			Description:     "Create a custom event site",
			FeatureGroupID:  groupIDs[m.FeatureGroupLandingPage],
			IsAvailable:     true,
			PackageTiersIDs: allIdsExceptFree,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "custom_business_page",
			Description:     "Create a custom business page to display your organization's events",
			FeatureGroupID:  groupIDs[m.FeatureGroupLandingPage],
			IsAvailable:     true,
			PackageTiersIDs: []uint{tierIDs[m.PackageTierNetwork]},
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "custom_emails",
			Description:     "Customize the emails sent to attendees",
			FeatureGroupID:  groupIDs[m.FeatureGroupEmailManagement],
			IsAvailable:     true,
			PackageTiersIDs: allIdsExceptFree,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
		{
			Name:            "max_ticket_releases_per_event",
			Description:     "Maximum number of ticket releases per event",
			FeatureGroupID:  groupIDs[m.FeatureGroupTicketManagement],
			IsAvailable:     true,
			PackageTiersIDs: allIds,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree], Limit: pID(1), LimitDescription: "1"},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent], Limit: pID(2), LimitDescription: "2"},
				{PackageTierID: tierIDs[m.PackageTierProfessional], Limit: pID(10), LimitDescription: "10"},
				{PackageTierID: tierIDs[m.PackageTierNetwork], LimitDescription: "Unlimited"},
			},
		},
		{
			Name:            "max_ticket_types_per_ticket_release",
			Description:     "Maximum number of ticket types per event",
			FeatureGroupID:  groupIDs[m.FeatureGroupTicketManagement],
			IsAvailable:     true,
			PackageTiersIDs: allIds,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree], Limit: pID(1), LimitDescription: "1"},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent], Limit: pID(5), LimitDescription: "5"},
				{PackageTierID: tierIDs[m.PackageTierProfessional], Limit: pID(10), LimitDescription: "10"},
				{PackageTierID: tierIDs[m.PackageTierNetwork], LimitDescription: "Unlimited"},
			},
		},
		{
			Name:            "max_ticket_types_per_ticket_release",
			Description:     "Maximum number of ticket types per event",
			FeatureGroupID:  groupIDs[m.FeatureGroupTicketManagement],
			IsAvailable:     true,
			PackageTiersIDs: allIds,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierFree], Limit: pID(1)},
				{PackageTierID: tierIDs[m.PackageTierSingleEvent], Limit: pID(5)},
				{PackageTierID: tierIDs[m.PackageTierProfessional], Limit: pID(10)},
				{PackageTierID: tierIDs[m.PackageTierNetwork], LimitDescription: "Unlimited"},
			},
		},
		{
			Name:            "reserved_ticket_releases",
			Description:     "Reserved ticket releases for events",
			FeatureGroupID:  groupIDs[m.FeatureGroupTicketManagement],
			IsAvailable:     true,
			PackageTiersIDs: allIdsExceptFree,
			FeatureLimits: []m.FeatureLimit{
				{PackageTierID: tierIDs[m.PackageTierSingleEvent]},
				{PackageTierID: tierIDs[m.PackageTierProfessional]},
				{PackageTierID: tierIDs[m.PackageTierNetwork]},
			},
		},
	}

	return features
}

func addFreeTierToFeature(feature m.Feature, freeTierID uint) m.Feature {
	// Check if the Free tier is already included in the PackageTiersIDs
	included := false
	for _, id := range feature.PackageTiersIDs {
		if id == freeTierID {
			included = true
			break
		}
	}
	// If not included, add the Free tier to PackageTiersIDs
	if !included {
		feature.PackageTiersIDs = append(feature.PackageTiersIDs, freeTierID)
	}
	// Ensure the Free tier has a corresponding FeatureLimit if applicable
	freeTierLimitIncluded := false
	for _, limit := range feature.FeatureLimits {
		if limit.PackageTierID == freeTierID {
			freeTierLimitIncluded = true
			break
		}
	}
	// Add a default FeatureLimit for the Free tier if not already included
	if !freeTierLimitIncluded {
		feature.FeatureLimits = append(feature.FeatureLimits, m.FeatureLimit{
			PackageTierID: freeTierID,
		})
	}
	return feature
}

func FreeBetaFeatures(db *gorm.DB) []m.Feature {
	_, tierIDs := fetchIDsForGroupsAndTiers(db)
	freeTierID := tierIDs[m.PackageTierFree]

	// Get the default features
	defaultFeatures := DefaultFeatures(db)

	// Create FreeBetaFeatures by adding Free tier to each feature
	freeBetaFeatures := make([]m.Feature, len(defaultFeatures))
	for i, feature := range defaultFeatures {
		freeBetaFeatures[i] = addFreeTierToFeature(feature, freeTierID)
	}

	return freeBetaFeatures
}

// Function that initially creates the default features in the database if they dont exist
// NOTE: This function is called in main.go and should not be called again in the application
// After the initial creation, the features can be updated in the database
// Use react-admin to update the features in the database
func InitializeDefaultFeatures(db *gorm.DB) error {
	features := DefaultFeatures(db)

	for _, feature := range features {
		var existingFeature m.Feature
		db.Where("name = ?", feature.Name).First(&existingFeature)

		if existingFeature.ID != 0 {
			if os.Getenv("ENV") == "prod" {
				// return since we dont want to update the features in production
				return nil
			}
			continue
		}

		featureToBeCreated := m.Feature{
			Name:            feature.Name,
			Description:     feature.Description,
			FeatureGroupID:  feature.FeatureGroupID,
			IsAvailable:     feature.IsAvailable,
			PackageTiersIDs: feature.PackageTiersIDs,
		}

		if existingFeature.ID == 0 {
			// Create it
			if err := db.Create(&featureToBeCreated).Error; err != nil {
				log.Println(err)
				return err
			}
		}

		// Relate the feature to the package tiers using the PackageTiersIDs that isnt stored in the database

		for _, tierID := range feature.PackageTiersIDs {
			var packageTier m.PackageTier
			if err := db.First(&packageTier, tierID).Error; err != nil {
				log.Println(err)
				return err
			}

			featureToBeCreated.PackageTiers = append(featureToBeCreated.PackageTiers, packageTier)
		}

		var featureLimits []m.FeatureLimit

		for _, featureLimit := range feature.FeatureLimits {
			featureLimit.FeatureID = featureToBeCreated.ID
			featureLimits = append(featureLimits, featureLimit)
		}

		featureToBeCreated.FeatureLimits = featureLimits

		// Save it
		if err := db.Save(&featureToBeCreated).Error; err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

// Function that initially creates the default features in the database if they don't exist
// NOTE: This function is called in main.go and should not be called again in the application
// After the initial creation, the features can be updated in the database
// Use react-admin to update the features in the database
func InitializeBETADefaultFeatures(db *gorm.DB) error {
	features := FreeBetaFeatures(db) // Use FreeBetaFeatures instead of DefaultFeatures

	for _, feature := range features {
		var existingFeature m.Feature
		db.Where("name = ?", feature.Name).First(&existingFeature)

		if existingFeature.ID != 0 {
			if os.Getenv("ENV") == "prod" {
				// return since we don't want to update the features in production
				return nil
			}
			continue
		}

		featureToBeCreated := m.Feature{
			Name:            feature.Name,
			Description:     feature.Description,
			FeatureGroupID:  feature.FeatureGroupID,
			IsAvailable:     feature.IsAvailable,
			PackageTiersIDs: feature.PackageTiersIDs,
		}

		if existingFeature.ID == 0 {
			// Create it
			if err := db.Create(&featureToBeCreated).Error; err != nil {
				log.Println(err)
				return err
			}
		}

		// Relate the feature to the package tiers using the PackageTiersIDs that isn't stored in the database
		for _, tierID := range feature.PackageTiersIDs {
			var packageTier m.PackageTier
			if err := db.First(&packageTier, tierID).Error; err != nil {
				log.Println(err)
				return err
			}

			featureToBeCreated.PackageTiers = append(featureToBeCreated.PackageTiers, packageTier)
		}

		var featureLimits []m.FeatureLimit

		for _, featureLimit := range feature.FeatureLimits {
			featureLimit.FeatureID = featureToBeCreated.ID
			featureLimits = append(featureLimits, featureLimit)
		}

		featureToBeCreated.FeatureLimits = featureLimits

		// Save it
		if err := db.Save(&featureToBeCreated).Error; err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
