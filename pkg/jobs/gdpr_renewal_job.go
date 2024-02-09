package jobs

import (
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var gdpr_logger = logrus.New()
var gdpr_logger_file string = "logs/gdpr_renewal_job.log"

func init() {
	// Load or create log file
	// Create logs directory if it doesn't exist
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	if _, err := os.Stat(gdpr_logger_file); os.IsNotExist(err) {
		os.Create(gdpr_logger_file)
	}

	file, err := os.OpenFile(gdpr_logger_file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		gdpr_logger.Fatal(err)
	}

	gdpr_logger.SetFormatter(&logrus.JSONFormatter{})
	gdpr_logger.SetOutput(file)
	gdpr_logger.SetLevel(logrus.DebugLevel)
}

func GDPRRenewalNotifyJob(db *gorm.DB) error {
	gdpr_logger.Info("Starting GDPR renewal notify job")

	var users []models.User
	var err error
	if err := db.
		Preload("FoodPreferences").
		Find(&users).Error; err != nil {
		gdpr_logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error fetching users from the database")

		return err
	}

	// Send an email to the user asking them to renew their GDPR agreement.
	// If they don't renew their GDPR agreement within a month, delete their record from the database.
	for _, user := range users {
		foodPreferences := user.FoodPreferences
		if foodPreferences.GDPRAgreed {
			foodPreferences.NeedsToRenewGDPR = true

			if err := db.Save(&foodPreferences).Error; err != nil {
				gdpr_logger.WithFields(logrus.Fields{
					"error": err,
					"user":  user.UGKthID,
				}).Error("Error saving user food preference in the database")

				return err
			}
			err = Notify_GDPRFoodPreferencesRenewal(db, &user)
			if err != nil {
				gdpr_logger.WithFields(logrus.Fields{
					"error": err,
					"user":  user.UGKthID,
				}).Error("Error sending GDPR renewal email to user")

				continue
			}
		}
	}

	gdpr_logger.Info("GDPR renewal notify job completed")

	return nil
}

func GDPRCheckRenewalJob(db *gorm.DB) error {
	gdpr_logger.Info("Starting GDPR renewal check job")

	var users []models.User
	if err := db.
		Preload("FoodPreferences").
		Find(&users).Error; err != nil {
		gdpr_logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error fetching users from the database")
		return err
	}

	// Check if the user has renewed their GDPR agreement within a month.
	// If they haven't, delete their record from the database.
	for _, user := range users {
		foodPreferences := user.FoodPreferences
		if foodPreferences.NeedsToRenewGDPR {
			// Hard delete the user food preference
			if err := db.Unscoped().Where("user_ug_kth_id = ?", user.UGKthID).Delete(&models.UserFoodPreference{}).Error; err != nil {
				gdpr_logger.WithFields(logrus.Fields{
					"error": err,
					"user":  user.UGKthID,
				}).Error("Error deleting user food preference from the database")

				return err
			}

			var userFoodPreference models.UserFoodPreference = models.UserFoodPreference{
				UserUGKthID:      user.UGKthID,
				NeedsToRenewGDPR: false,
			}

			// Create the clean row
			if err := db.Create(&userFoodPreference).Error; err != nil {
				gdpr_logger.WithFields(logrus.Fields{
					"error": err,
					"user":  user.UGKthID,
				}).Error("Error creating user food preference in the database")

				return err
			}
		} else {
			// If the user has renewed their GDPR agreement, set the RenewGDPRAgreed field to false.
			foodPreferences.GDPRAgreed = true
			if err := db.Save(&foodPreferences).Error; err != nil {
				gdpr_logger.WithFields(logrus.Fields{
					"error": err,
					"user":  user.UGKthID,
				}).Error("Error saving user food preference in the database")

				return err
			}
		}
	}

	gdpr_logger.Info("GDPR renewal check job completed")

	return nil
}
