package services

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type EventLandingPageService struct {
	db *gorm.DB
}

func NewEventLandingPageService(db *gorm.DB) *EventLandingPageService {
	return &EventLandingPageService{db: db}
}

func (elp *EventLandingPageService) SaveEventLandingPage(body *models.EventLandingPage) *types.ErrorResponse {
	err := body.Validate()
	if err != nil {
		return &types.ErrorResponse{StatusCode: 400, Message: err.Error()}
	}

	// Find existing landing page
	var existing models.EventLandingPage
	result := elp.db.Where("event_id = ?", body.EventID).First(&existing)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return &types.ErrorResponse{StatusCode: 500, Message: result.Error.Error()}
	}

	// If landing page exists, update it
	if existing.ID != 0 {
		result = elp.db.Model(&existing).Updates(body)
		if result.Error != nil {
			return &types.ErrorResponse{StatusCode: 500, Message: result.Error.Error()}
		}
	} else {
		// If landing page does not exist, create it
		result = elp.db.Create(body)
		if result.Error != nil {
			return &types.ErrorResponse{StatusCode: 500, Message: result.Error.Error()}
		}
	}

	return nil
}

func (elp *EventLandingPageService) SaveEventLandingPageEditorState(stateBytes []byte, eventID uint) *types.ErrorResponse {
	// Find existing landing page
	var existing models.EventLandingPage
	result := elp.db.Where("event_id = ?", eventID).First(&existing)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return &types.ErrorResponse{StatusCode: 500, Message: result.Error.Error()}
	}

	// If landing page exists, update it
	if existing.ID != 0 {
		result = elp.db.Model(&existing).Update("editor_state", stateBytes)
		if result.Error != nil {
			return &types.ErrorResponse{StatusCode: 500, Message: result.Error.Error()}
		}
	} else {
		// If landing page does not exist, create it
		landingPage := models.EventLandingPage{
			EventID:     eventID,
			EditorState: stateBytes,
		}
		result = elp.db.Create(&landingPage)
		if result.Error != nil {
			return &types.ErrorResponse{StatusCode: 500, Message: result.Error.Error()}
		}
	}

	return nil
}
