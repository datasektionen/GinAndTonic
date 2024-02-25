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

func (os *OrganisationService) AddUserToOrganization(username string, organizationID uint, organizationRole models.OrgRole) error {
	var user models.User
	var organization models.Organization

	if err := os.DB.First(&user, "username = ?", username).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	if err := os.DB.First(&organization, organizationID).Error; err != nil {
		return fmt.Errorf("organization not found")
	}

	// Start transaction
	tx := os.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Associate user with organization
	if err := tx.Model(&organization).Association("Users").Append(&user); err != nil {
		tx.Rollback()
		return err
	}
	// 2. Create organization user role
	organizationUserRole := models.OrganizationUserRole{
		UserUGKthID:          user.UGKthID,
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

func (os *OrganisationService) RemoveUserFromOrganization(username string, organizationID uint) error {
	var user models.User
	var organization models.Organization

	if err := os.DB.First(&user, "username = ?", username).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	if err := os.DB.Preload("Users").Preload("Users.OrganizationUserRoles").First(&organization, organizationID).Error; err != nil {
		return fmt.Errorf("organization not found")
	}

	isOwner, err := os.isUserOwnerOfOrganization(user.UGKthID, organization)
	if err != nil {
		return err
	}

	if isOwner && len(organization.Users) == 1 {
		return fmt.Errorf("user %v is the owner of the organization %v", username, organizationID)
	}

	if err := os.DB.Model(&organization).Association("Users").Delete(&user); err != nil {
		return fmt.Errorf("there was an error removing the user from the organization: %w", err)
	}

	// Remove the user.OrganizationUserRole for this organization
	if err := os.DB.Unscoped().Where("user_ug_kth_id = ? AND organization_id = ?", user.UGKthID, organization.ID).Delete(&models.OrganizationUserRole{}).Error; err != nil {
		return fmt.Errorf("there was an error removing the user from the organization: %w", err)
	}

	return nil
}

func (os *OrganisationService) GetOrganizationUsers(organizationID uint) ([]models.User, error) {
	var organization models.Organization

	if err := os.DB.First(&organization, organizationID).Error; err != nil {
		return nil, fmt.Errorf("organization not found")
	}

	users, err := organization.GetUsers(os.DB)

	if err != nil {
		return nil, fmt.Errorf("there was an error fetching the organization users: %w", err)
	}

	return users, nil
}

func (os *OrganisationService) isUserOwnerOfOrganization(userUGKthID string, organization models.Organization) (bool, error) {
	orgOwners, err := models.GetOrganizationOwners(os.DB, organization)
	if err != nil {
		return false, fmt.Errorf("there was an error fetching the organization owners: %w", err)
	}

	for _, owner := range orgOwners {
		if owner.UGKthID == userUGKthID {
			return true, nil
		}
	}

	return false, nil
}

func (os *OrganisationService) ChangeUserRoleInOrganization(username string, organizationID uint, newRole models.OrgRole) error {
	var user models.User
	var organization models.Organization

	// Find the user
	if err := os.DB.First(&user, "username = ?", username).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	// Find the organization
	if err := os.DB.First(&organization, organizationID).Error; err != nil {
		return fmt.Errorf("organization not found")
	}

	// Find the organization user role
	var organizationUserRole models.OrganizationUserRole
	if err := os.DB.Where("user_ug_kth_id = ? AND organization_id = ?", user.UGKthID, organization.ID).First(&organizationUserRole).Error; err != nil {
		return fmt.Errorf("organization user role not found")
	}

	// Start transaction
	tx := os.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the role
	if err := tx.Model(&organizationUserRole).Update("organization_role_name", string(newRole)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("there was an error updating the user role: %w", err)
	}

	// Commit transaction
	tx.Commit()

	return nil
}
