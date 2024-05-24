package manager_controller

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	manager_service "github.com/DowLucas/gin-ticket-release/pkg/services/manager"
	surfboard_service_merchant "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/merchant"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ManagerMerchantController struct {
	DB              *gorm.DB
	network_service *manager_service.NetworkService
}

func NewManagerMerchantController(db *gorm.DB) *ManagerMerchantController {
	return &ManagerMerchantController{
		DB:              db,
		network_service: manager_service.NewNetworkService(db),
	}
}

func (mmc *ManagerMerchantController) CreateNetworkMerchant(c *gin.Context) {
	var formValues types.MerchantBusinessData
	if err := c.ShouldBindJSON(&formValues); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx := mmc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	user := c.MustGet("user").(models.User)

	network, rerr := mmc.network_service.GetNetworkDetails(&user)
	if rerr != nil {
		c.JSON(rerr.StatusCode, gin.H{"error": rerr.Message})
		return
	}

	if network.Merchant.HasOngoingApplication() {
		c.JSON(400, gin.H{"error": "There is an ongoing application"})
		return
	}

	// Update network details
	network.Details.LegalName = formValues.LegalName
	network.Details.CorporateID = formValues.CorporateID
	network.Details.LegalName = formValues.LegalName
	network.Details.CountryCode = formValues.CountryCode
	network.Details.Email = formValues.BusinessEmail
	network.Details.AddressLine1 = formValues.AddressLine1
	network.Details.AddressLine2 = formValues.AddressLine2
	network.Details.City = formValues.City
	network.Details.PostalCode = formValues.PostalCode
	network.Details.PhoneCode = formValues.PhoneCode
	network.Details.PhoneNumber = formValues.PhoneNumber
	network.Details.StoreName = formValues.StoreName

	if err := tx.Save(&network).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err := surfboard_service_merchant.CreateMerchant(tx, &user, network)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(201, gin.H{"message": "Merchant created"})
}
