package manager_service

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type ManagerService struct {
	DB *gorm.DB
}

func NewManagerService(db *gorm.DB) *ManagerService {
	return &ManagerService{
		DB: db,
	}
}

func (ms *ManagerService) GetNetworkEvents(user *models.User) (allEvents []models.Event, rerr *types.ErrorResponse) {
	// If the user belongs to a network, get all events for that network.
	if user.NetworkID != nil {
		var events []models.Event

		for _, organization := range user.Organizations {
			if err := ms.DB.Where("organization_id = ?", organization.ID).Find(&events).Error; err != nil {
				return nil, &types.ErrorResponse{
					StatusCode: 500,
					Message:    "Unable to get organization events",
				}
			}

			allEvents = append(allEvents, events...)
		}
	}

	return allEvents, nil
}
