package merchant_job

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	surfboard_service_merchant "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/merchant"
	"gorm.io/gorm"
)

// Job that runs every 30 minutes to update all merchants statuses
func UpdateMerchantStatuses(db *gorm.DB) {
	// Get all merchants
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	merchants, err := models.GetAllOngoingApplications(tx)
	if err != nil {
		tx.Rollback()
		return
	}


	// Update statuses
	for _, merchant := range merchants {
		err := surfboard_service_merchant.CheckApplicationStatus(tx, &merchant)
		if err != nil {
			tx.Rollback()
			return
		}
	}

	tx.Commit()
}
	
