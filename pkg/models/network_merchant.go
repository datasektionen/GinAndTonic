package models

import (
	"gorm.io/gorm"
)

type SurfApplicationStatus string

const (
	APPLICATION_NOT_STARTED         SurfApplicationStatus = "application_not_started"
	APPLICATION_INITIATED           SurfApplicationStatus = "application_initiated"
	APPLICATION_SUBMITTED           SurfApplicationStatus = "application_submitted"
	APPLICATION_PENDING_INFORMATION SurfApplicationStatus = "application_pending_information"
	APPLICATION_SIGNED              SurfApplicationStatus = "application_signed"
	APPLICATION_REJECTED            SurfApplicationStatus = "application_rejected"
	APPLICATION_COMPLETED           SurfApplicationStatus = "application_completed"
	APPLICATION_EXPIRED             SurfApplicationStatus = "application_expired"
	MERCHANT_CREATED                SurfApplicationStatus = "merchant_created"
)

// NetworkMerchant represents a network merchant in the system.
// swagger:model
type NetworkMerchant struct {
	gorm.Model
	// the ID of the network associated with the merchant
	// required: true
	NetworkID uint `json:"network_id" gorm:"uniqueIndex:idx_network_merchant_network_id_merchant_id"`
	// the ID of the application associated with the merchant
	// required: true
	ApplicationID string `json:"applicationId"`
	// the unique identifier for the merchant
	// required: true
	MerchantID string `json:"merchantId"`
	// the current status of the merchant's application
	// required: true
	// enum: application_initiated,application_submitted,application_pending_information,application_signed,application_rejected,application_completed,application_expired,merchant_created
	ApplicationStatus SurfApplicationStatus `json:"applicationStatus" default:"application_not_started"`
	// the URL for the merchant to fill in necessary information
	// required: true
	WebKybURL string `json:"webKybUrl"`
	// the unique identifier for the merchant's store
	// required: true
	StoreID string `json:"storeId"`
}

func (nm NetworkMerchant) IsApplicationCompleted() bool {
	return nm.ApplicationStatus == MERCHANT_CREATED
}

func (nm NetworkMerchant) HasOngoingApplication() bool {
	if nm.ID == 0 {
		return false
	}

	if nm.ApplicationStatus == APPLICATION_COMPLETED {
		return false
	}

	return nm.ApplicationStatus != APPLICATION_REJECTED && nm.ApplicationStatus != APPLICATION_EXPIRED
}

func GetAllOngoingApplications(db *gorm.DB) ([]NetworkMerchant, error) {
	var merchants []NetworkMerchant
	// Should return all NetworkMerchants where ApplicationStatus is not MERCHANT_CREATED or APPLICATION_EXPIRED or APPLICATION_REJECTED
	if err := db.Where("application_status != ? AND application_status != ? AND application_status != ?", MERCHANT_CREATED, APPLICATION_EXPIRED, APPLICATION_REJECTED).Find(&merchants).Error; err != nil {
		return merchants, err
	}
	return merchants, nil
}

func GetAllMerchants(db *gorm.DB) ([]NetworkMerchant, error) {
	var merchants []NetworkMerchant
	if err := db.Find(&merchants).Error; err != nil {
		return merchants, err
	}
	return merchants, nil
}
