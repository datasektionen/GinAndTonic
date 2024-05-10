package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type FeatureUsage struct {
	CreatedAt        time.Time `gorm:"primaryKey"`
	FeatureID        uint      `gorm:"primaryKey;autoIncrement:false" json:"feature_id"`
	PlanEnrollmentID uint      `gorm:"primaryKey;autoIncrement:false" json:"plan_enrollment_id"`
	ObjectReference  *string   `gorm:"default:null" json:"object_reference"`
	Usage            int       `json:"usage"`
}

/*
Outline of how FeatureUsage works

- Each entry in the FeatureUsage table is a record of how many times a feature has been used by a user, timestamped by the time of creation
- The primary key is a composite key of the feature ID and the plan enrollment ID
- The usage field is an integer that represents the number of times the feature has been used
*/
func GetLatestFeatureUsage(db *gorm.DB, featureID, planEnrollmentID uint, objectReference *string) (featureUsage *FeatureUsage, err error) {
	query := db.Where("feature_id = ? AND plan_enrollment_id = ?", featureID, planEnrollmentID)
	if objectReference != nil {
		query = query.Where("object_reference = ?", *objectReference)
	}
	if err := query.Order("created_at desc").First(&featureUsage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}
	return featureUsage, nil
}

/*
- Get the feature usage data for the last 30 days
*/
func GetMonthlyFeatureUsages(db *gorm.DB, featureID, planEnrollmentID uint, objectReference *string) (featureUsages []FeatureUsage, err error) {
	query := db.Where("feature_id = ? AND plan_enrollment_id = ? AND created_at >= ?", featureID, planEnrollmentID, time.Now().AddDate(0, 0, -30))
	if objectReference != nil {
		query = query.Where("object_reference = ?", *objectReference)
	}
	if err := query.Find(&featureUsages).Error; err != nil {
		return nil, err
	}
	return featureUsages, nil
}

/*
- Get the feature usage data for the last year
*/
func GetYearlyFeatureUsages(db *gorm.DB, featureID, planEnrollmentID uint, objectReference *string) (featureUsage []FeatureUsage, err error) {
	query := db.Where("feature_id = ? AND plan_enrollment_id = ? AND created_at >= ?", featureID, planEnrollmentID, time.Now().AddDate(-1, 0, 0))
	if objectReference != nil {
		query = query.Where("object_reference = ?", *objectReference)
	}
	if err := query.Find(&featureUsage).Error; err != nil {
		return nil, err
	}
	return featureUsage, nil
}

/*
- Given a list of FeatureUsage objects, return the usage from the first object to the last object
- Take the final usage and subtract the initial usage to get the total usage
*/
func GetTotalFeatureUsage(usage []FeatureUsage) (totalUsage int) {
	if len(usage) == 0 {
		return 0
	}

	// If there's only one entry, return 0
	if len(usage) == 1 {
		return usage[0].Usage
	}

	// Assert that the usage list is ordered by time
	if usage[0].CreatedAt.After(usage[len(usage)-1].CreatedAt) {
		return 0
	}

	// Calculate total usage
	totalUsage = usage[len(usage)-1].Usage - usage[0].Usage

	// If usage has decreased, return 0
	if totalUsage < 0 {
		return 0
	}

	return totalUsage
}
