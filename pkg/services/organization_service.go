package services

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

type OrganisationService struct {
	DB *gorm.DB
}

func NewOrganizationService(db *gorm.DB) *OrganisationService {
	return &OrganisationService{DB: db}
}

func (os *OrganisationService) AddUserToOrganization(userUGKthID string, organizationID uint, organizationRole models.OrgRole) error {
	var user models.User
	var organization models.Organization

	if err := os.DB.First(&user, "ug_kth_id = ?", userUGKthID).Error; err != nil {
		return fmt.Errorf("User not found")
	}

	if err := os.DB.First(&organization, organizationID).Error; err != nil {
		return fmt.Errorf("Organization not found")
	}

	// Start transaction
	tx := os.DB.Begin()
	// 1. Associate user with organization
	if err := tx.Model(&organization).Association("Users").Append(&user); err != nil {
		tx.Rollback()
		return err
	}
	// 2. Create organization user role
	organizationUserRole := models.OrganizationUserRole{
		UserUGKthID:          userUGKthID,
		OrganizationID:       organization.ID,
		OrganizationRoleName: string(organizationRole),
	}

	if err := tx.Create(&organizationUserRole).Error; err != nil {
		tx.Rollback()
		return err
	}
	// Commit transaction
	tx.Commit()

	return nil
}

func (os *OrganisationService) RemoveUserFromOrganization(userUGKthID string, organizationID uint) error {
	var user models.User
	var organization models.Organization

	if err := os.DB.First(&user, "ug_kth_id = ?", userUGKthID).Error; err != nil {
		return fmt.Errorf("User not found")
	}

	if err := os.DB.Preload("Users").Preload("Users.OrganizationUserRoles").First(&organization, organizationID).Error; err != nil {
		return fmt.Errorf("Organization not found")
	}

	isOwner, err := os.isUserOwnerOfOrganization(userUGKthID, organization)
	if err != nil {
		return err
	}
	if isOwner {
		return fmt.Errorf("User %v is the owner of the organization %v", userUGKthID, organizationID)
	}

	if err := os.DB.Model(&organization).Association("Users").Delete(&user); err != nil {
		return fmt.Errorf("There was an error removing the user from the organization: %w", err)
	}

	return nil
}

func (os *OrganisationService) GetOrganizationUsers(organizationID uint) ([]models.User, error) {
	var organization models.Organization

	if err := os.DB.First(&organization, organizationID).Error; err != nil {
		return nil, fmt.Errorf("Organization not found")
	}

	users, err := organization.GetUsers(os.DB)

	if err != nil {
		return nil, fmt.Errorf("There was an error fetching the organization users: %w", err)
	}

	return users, nil
}

func (os *OrganisationService) isUserOwnerOfOrganization(userUGKthID string, organization models.Organization) (bool, error) {
	orgOwners, err := models.GetOrganizationOwners(os.DB, organization)
	if err != nil {
		return false, fmt.Errorf("There was an error fetching the organization owners: %w", err)
	}

	for _, owner := range orgOwners {
		if owner.UGKthID == userUGKthID {
			return true, nil
		}
	}

	return false, nil
}
