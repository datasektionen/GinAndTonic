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
	if user.NetworkID == nil {
		return nil, nil
	}

	network, err := models.GetNetworkByID(ns.DB, *user.NetworkID)
	if err != nil {
		return nil, &types.ErrorResponse{
			StatusCode: 500,
			Message:    "Unable to get network details",
		}
	}

	var nur models.NetworkUserRole
	if err := ns.DB.Where("user_ug_kth_id = ? AND network_id = ?", user.UGKthID, user.NetworkID).First(&nur).Error; err != nil {
		return nil, &types.ErrorResponse{
			StatusCode: 500,
			Message:    "Unable to get network role",
		}
	}

	for _, organization := range network.Organizations {
		users, err := organization.GetUsers(ns.DB)
		if err != nil {
			return nil, &types.ErrorResponse{
				StatusCode: 500,
				Message:    "Unable to get organization users",
			}
		}

		organization.Users = users
	}

	// Handles rbac for network roles
	switch nur.NetworkRoleName {
	case models.NetworkMember:
		// Keep your own network role
		network.NetworkUserRoles = []models.NetworkUserRole{nur}
	}

	return network, nil

}
