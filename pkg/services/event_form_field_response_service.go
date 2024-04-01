package services

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type EventFormFieldResponseService struct {
	db *gorm.DB
}

func NewEventFormFieldResponseService(db *gorm.DB) *EventFormFieldResponseService {
	return &EventFormFieldResponseService{db: db}
}

func (s *EventFormFieldResponseService) Upsert(user *models.User,
	ticketRequestID string,
	responses []types.EventFormFieldResponseCreateRequest) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ticketRequest models.TicketRequest
	if err := tx.Where("id = ? AND user_ug_kth_id = ?", ticketRequestID, user.UGKthID).First(&ticketRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	var existingResponses []models.EventFormFieldResponse
	if err := tx.Preload("EventFormField").Where("ticket_request_id = ?", ticketRequestID).Find(&existingResponses).Error; err != nil {
		tx.Rollback()
		return err
	}

	existingResponseMap := make(map[uint]models.EventFormFieldResponse)
	for _, response := range existingResponses {
		existingResponseMap[response.EventFormFieldID] = response
	}

	for _, response := range responses {
		existingResponse, exists := existingResponseMap[response.EventFormFieldID]

		if exists {
			// Update the existing response
			existingResponse.Value = response.Value
			// Print the type of the response
			if err := tx.Model(models.EventFormFieldResponse{}).Where("id = ?", existingResponse.ID).UpdateColumn("value", response.Value).Error; err != nil {
				tx.Rollback()
				return err
			}
			// Remove the response from the map
			delete(existingResponseMap, response.EventFormFieldID)
		} else {
			var eventFormField models.EventFormField
			if err := tx.Where("id = ?", response.EventFormFieldID).First(&eventFormField).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// Handle not found error
					return fmt.Errorf("EventFormField with ID %d not found", response.EventFormFieldID)
				}
				// Handle other DB errors
				return err
			}

			// Create a new response
			newResponse := models.EventFormFieldResponse{
				TicketRequestID:  ticketRequest.ID,
				EventFormFieldID: response.EventFormFieldID,
				Value:            response.Value,
			}

			if err := tx.Create(&newResponse).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Delete the responses that are not included in the request
	for _, response := range existingResponseMap {
		if err := tx.Delete(&response).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
