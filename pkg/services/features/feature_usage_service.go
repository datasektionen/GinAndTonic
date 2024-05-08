package feature_services

import (
	"errors"
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func UpdateFeatureUsage(tx *gorm.DB, planEnrollmentID uint, featureName string, increment int, objectReference *string) error {
	if planEnrollmentID == 0 {
		return errors.New("plan enrollment ID is required when preloading features")
	}

	fmt.Println("Incrementing feature usage for feature", featureName)

	var feature models.Feature
	if err := tx.First(&feature, "name = ?", featureName).Error; err != nil {
		return errors.New("feature not found")
	}

	latestUsage, err := models.GetLatestFeatureUsage(tx, feature.ID, planEnrollmentID, objectReference)
	if err != nil {
		return err
	}

	if latestUsage == nil {
		// Insert initial usage with usage = 0
		var featureUsage models.FeatureUsage = models.FeatureUsage{
			FeatureID:        feature.ID,
			PlanEnrollmentID: planEnrollmentID,
			Usage:            0,
			ObjectReference:  objectReference,
		}

		if err := tx.Create(&featureUsage).Error; err != nil {
			return err
		}
	}

	var planEnrollment models.PlanEnrollment
	if err := tx.First(&planEnrollment, planEnrollmentID).Error; err != nil {
		return err
	}

	var usage int = 0
	if latestUsage != nil {
		usage = latestUsage.Usage
	}

	var featureUsage models.FeatureUsage = models.FeatureUsage{
		FeatureID:        feature.ID,
		PlanEnrollmentID: planEnrollmentID,
		Usage:            usage + increment,
		ObjectReference:  objectReference,
	}

	if err := tx.Create(&featureUsage).Error; err != nil {
		return err
	}

	return nil
}

func IncrementFeatureUsage(tx *gorm.DB, planEnrollmentID uint, featureName string, objectReference *string) error {
	return UpdateFeatureUsage(tx, planEnrollmentID, featureName, 1, objectReference)
}

// Function that takes a list of feature names and increments the usage of each feature by 1
func IncrementFeatureUsages(tx *gorm.DB, planEnrollmentID uint, featureNames []string, objectReference *string) error {
	for _, featureName := range featureNames {
		if err := UpdateFeatureUsage(tx, planEnrollmentID, featureName, 1, objectReference); err != nil {
			return err
		}
	}

	return nil
}
