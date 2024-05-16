package models

import (
	"gorm.io/gorm"
)

type SurfApplicationStatus string

const (
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
	NetworkID uint `json:"network_id"`
	// the ID of the application associated with the merchant
	// required: true
	ApplicationID string `json:"applicationId"`
	// the unique identifier for the merchant
	// required: true
	MerchantID string `json:"merchantId"`
	// the current status of the merchant's application
	// required: true
	// enum: application_initiated,application_submitted,application_pending_information,application_signed,application_rejected,application_completed,application_expired,merchant_created
	ApplicationStatus SurfApplicationStatus `json:"applicationStatus"`
	// the URL for the merchant to fill in necessary information
	// required: true
	WebKybURL string `json:"webKybUrl"`
	// the unique identifier for the merchant's store
	// required: true
	StoreID string `json:"storeId"`
}
