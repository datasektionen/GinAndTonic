package services

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type BankingService struct {
	DB *gorm.DB
}

func NewBankingService(db *gorm.DB) *BankingService {
	return &BankingService{
		DB: db,
	}
}

func (s *BankingService) SubmitBankingDetails(bdr *types.BankingDetailsRequest, orgId uint) (rerr *types.ErrorResponse) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var organization models.Organization
	err := tx.First(&organization, orgId).Error
	if err != nil {
		return &types.ErrorResponse{
			StatusCode: 404,
			Message:    "Organization not found",
		}
	}

	b := models.BankingDetail{
		OrganizationID: orgId,
		BankName:       bdr.BankName,
		AccountHolder:  bdr.AccountHolder,
		AccountNumber:  bdr.AccountNumber,
		ClearingNumber: bdr.ClearingNumber,
	}

	if err := b.Validate(); err != nil {
		return &types.ErrorResponse{
			StatusCode: 400,
			Message:    err.Error(),
		}
	}

	err = b.EncryptFields()

	if err != nil {
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    "Failed to encrypt banking details",
		}
	}

	var existingBankingDetail models.BankingDetail
	if err := tx.Where("organization_id = ?", orgId).First(&existingBankingDetail).Error; err != nil {
		if err := tx.Create(&b).Error; err != nil {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 500,
				Message:    "Failed to create banking details",
			}
		}
	} else {
		if err := tx.Model(&existingBankingDetail).Updates(b).Error; err != nil {
			tx.Rollback()
			return &types.ErrorResponse{
				StatusCode: 500,
				Message:    "Failed to update banking details",
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    "Failed to commit transaction",
		}
	}

	return nil
}

func (s *BankingService) GetBankingDetails(orgId uint) (bd models.BankingDetail, rerr *types.ErrorResponse) {
	var bankingDetail models.BankingDetail
	err := s.DB.Where("organization_id = ?", orgId).First(&bankingDetail).Error
	if err != nil {
		return models.BankingDetail{}, &types.ErrorResponse{
			StatusCode: 404,
			Message:    "Banking details not found",
		}
	}

	err = bankingDetail.DecryptFields()
	if err != nil {
		return models.BankingDetail{}, &types.ErrorResponse{
			StatusCode: 500,
			Message:    "Failed to decrypt banking details",
		}
	}

	return bankingDetail, nil
}

func (s *BankingService) DeleteBankingDetails(orgId uint) (rerr *types.ErrorResponse) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var bankingDetail models.BankingDetail
	err := tx.Where("organization_id = ?", orgId).First(&bankingDetail).Error
	if err != nil {
		return &types.ErrorResponse{
			StatusCode: 404,
			Message:    "Banking details not found",
		}
	}

	// Important that we use Unscoped here to delete the record even if it's soft deleted
	if err := tx.Unscoped().Delete(&bankingDetail).Error; err != nil {
		tx.Rollback()
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    "Failed to delete banking details",
		}
	}

	if err := tx.Commit().Error; err != nil {
		return &types.ErrorResponse{
			StatusCode: 500,
			Message:    "Failed to commit transaction",
		}
	}

	return nil
}
