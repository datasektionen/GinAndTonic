package plan_enrollment_service

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type FreePlanEnrollmentService struct {
	DB *gorm.DB
}

// NewFreePlanEnrollmentService creates a new service with the given database client
func NewFreePlanEnrollmentService(db *gorm.DB) *FreePlanEnrollmentService {
	return &FreePlanEnrollmentService{DB: db}
}

func (fpes *FreePlanEnrollmentService) Enroll(user *models.User, body types.FreeEnrollmentPlanBody) *types.ErrorResponse {
	tx := fpes.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if the user is already enrolled in a plan
	plan, err := models.GetUserPlanEnrollment(tx, user)
	if err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error getting user plan enrollment"}
	}

	if plan.ID != 0 {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 400, Message: "User is already enrolled in a plan"}
	}

	// Get the package tier

	tier, err := models.GetPackageTierByType(fpes.DB, models.PackageTierFree)
	if err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error getting package tier"}
	}

	// check if org with same name already exists
	var eorg models.Organization
	if err := fpes.DB.Where("name = ?", body.Name).First(&eorg).Error; err == nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 400, Message: "Name is already taken"}
	}

	// Check if network with same name already exists
	var enetwork models.Network
	if err := fpes.DB.Where("name = ?", body.Name).First(&enetwork).Error; err == nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 400, Message: "Name is already taken"}
	}

	// Create the plan enrollment
	plan = models.PlanEnrollment{
		CreatorID:     user.UGKthID,
		OneTimePrice:  0,
		Plan:          models.NoPayment,
		PackageTierID: tier.ID,
		Features:      tier.DefaultFeatures,
	}

	if err := fpes.DB.Create(&plan).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error creating plan enrollment"}
	}

	// Create the organization
	org := models.Organization{
		Name:             body.Name,
		Email:            user.Email,
		PlanEnrollmentID: &plan.ID,
	}

	if err := fpes.DB.Create(&org).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error creating organization"}
	}

	// assign the user
	if org.AddUserWithRole(tx, *user, models.OrganizationOwner) != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error adding user to organization"}
	}

	referralSource := models.ReferralSource{
		UserUGKthID: user.UGKthID,
		Source:      body.ReferralSource,
		Specific:    body.ReferralSourceSpecific,
	}

	if err := fpes.DB.Create(&referralSource).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error creating referral source"}
	}

	tx.Commit()
	if tx.Error != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error committing transaction"}
	}

	return nil
}
