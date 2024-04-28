package admin_services

import "gorm.io/gorm"

type PricingPackageAdminService struct {
	DB *gorm.DB
}

// NewPricingPackageAdminService creates a new service with the given database client
func NewPricingPackageAdminService(db *gorm.DB) *PricingPackageAdminService {
	return &PricingPackageAdminService{DB: db}
}
