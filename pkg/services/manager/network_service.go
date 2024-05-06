package manager_service

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type NetworkService struct {
	DB *gorm.DB
}

func NewNetworkService(db *gorm.DB) *NetworkService {
	return &NetworkService{
		DB: db,
	}
}

func (ns *NetworkService) GetNetworkDetails(user *models.User) (network *models.Network, rerr *types.ErrorResponse) {
	// If the user belongs to a network, get the network details.
	if user.NetworkID != nil {
		network, err := models.GetNetworkByID(ns.DB, *user.NetworkID)
		if err != nil {
			return nil, &types.ErrorResponse{
				StatusCode: 500,
				Message:    "Unable to get network details",
			}
		}
		return &network, nil
	}

	return nil, nil
}
