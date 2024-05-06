package plan_enrollment_service

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

// FreePlanEnrollmentService is a service that handles the enrollment of users to the free plan.
type FreePlanEnrollmentService struct {
	DB *gorm.DB
}

// NewFreePlanEnrollmentService creates a new service with the given database client.
func NewFreePlanEnrollmentService(db *gorm.DB) *FreePlanEnrollmentService {
	return &FreePlanEnrollmentService{DB: db}
}

// Enroll enrolls a user to the free plan. It takes a User and a FreeEnrollmentPlanBody as arguments.
// It starts a database transaction and rolls back if any error occurs during the process.
// It checks if the user is already enrolled in a plan, if the organization or network name already exists,
// and creates the plan enrollment, network, organization, and referral source.
// If all operations are successful, it commits the transaction.
// It returns an ErrorResponse if any error occurs, otherwise it returns nil.
func (fpes *FreePlanEnrollmentService) Enroll(user *models.User, body types.FreeEnrollmentPlanBody) *types.ErrorResponse {
	// Start a database transaction
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

	// If the user is already enrolled in a plan, roll back the transaction and return an error
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

	// Check if an organization with the same name already exists
	var eorg models.Organization
	if err := fpes.DB.Where("name = ?", body.Name).First(&eorg).Error; err == nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 400, Message: "Name is already taken"}
	}

	// Check if a network with the same name already exists
	var enetwork models.Network
	if err := fpes.DB.Where("name = ?", body.Name).First(&enetwork).Error; err == nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 400, Message: "Name is already taken"}
	}

	// Get the default features that come with the free plan
	defaultFeatures, err := tier.GetDefaultFeatures(fpes.DB)
	if err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error getting default features"}
	}

	// Create the plan enrollment
	plan = models.PlanEnrollment{
		ReferenceName: tier.Name + "-" + user.UGKthID,
		CreatorID:     user.UGKthID,
		OneTimePrice:  0,
		Plan:          models.NoPayment,
		PackageTierID: tier.ID,
		Features:      defaultFeatures,
	}

	// If there's an error creating the plan enrollment, roll back the transaction and return an error
	if err := fpes.DB.Create(&plan).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error creating plan enrollment"}
	}

	// Create the network
	network := models.Network{
		Name:             body.Name,
		PlanEnrollmentID: plan.ID,
	}

	// If there's an error creating the network, roll back the transaction and return an error
	if err := fpes.DB.Create(&network).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error creating network"}
	}

	err = network.AddUserToNetwork(tx, *user, models.NetworkSuperAdmin)
	if err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error adding user to network"}
	}

	// Create the organization
	org := models.Organization{
		Name:      body.Name,
		Email:     user.Email,
		NetworkID: network.ID,
	}

	// If there's an error creating the organization, roll back the transaction and return an error
	if err := fpes.DB.Create(&org).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error creating organization"}
	}

	// Assign the user to the organization
	if org.AddUserWithRole(tx, *user, models.OrganizationOwner) != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error adding user to organization"}
	}

	// Create the referral source
	referralSource := models.ReferralSource{
		UserUGKthID: user.UGKthID,
		Source:      body.ReferralSource,
		Specific:    body.ReferralSourceSpecific,
	}

	// If there's an error creating the referral source, roll back the transaction and return an error
	if err := fpes.DB.Create(&referralSource).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error creating referral source"}
	}

	// Commit the transaction
	tx.Commit()
	if tx.Error != nil {
		tx.Rollback()
		return &types.ErrorResponse{StatusCode: 500, Message: "Error committing transaction"}
	}

	return nil
}
