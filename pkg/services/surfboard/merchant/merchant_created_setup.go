package surfboard_service_merchant

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	surfboard_service_store "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/store"
	"gorm.io/gorm"
)

func SetupMerchant(tx *gorm.DB, networkMerchant *models.NetworkMerchant) error {
	// SetupMerchant is a functio called when a merchant has been created and onborading is complete
	// It goes through the Networks organization and sets up a store for each organization

	var network models.Network
	err := tx.
		Preload("Organizations").
		Preload("Merchant").
		Preload("Details").
		Where("id = ?", networkMerchant.NetworkID).First(&network).Error
	if err != nil {
		return err
	}

	// Go through all organizations in the network and create a store for each
	for _, organization := range network.Organizations {
		if err := surfboard_service_store.CreateStore(&network, &organization, tx); err != nil {
			return err
		}
	}

	return nil
}
