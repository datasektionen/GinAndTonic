package admin_services

import "gorm.io/gorm"

type PlanEnrollmentAdminService struct {
	DB *gorm.DB
}

// NewPlanEnrollmentAdminService creates a new service with the given database client
func NewPlanEnrollmentAdminService(db *gorm.DB) *PlanEnrollmentAdminService {
	return &PlanEnrollmentAdminService{DB: db}
}
