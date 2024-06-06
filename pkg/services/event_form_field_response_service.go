package services

import (
	"errors"
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
	ticketID string,
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

	var ticket models.Ticket
	if err := tx.Where("id = ? AND user_ug_kth_id = ?", ticketID, user.UGKthID).First(&ticket).Error; err != nil {
		tx.Rollback()
		return errors.New("error getting ticket request")
	}

	var existingResponses []models.EventFormFieldResponse
	if err := tx.Preload("EventFormField").Where("ticket_id = ?", ticketID).Find(&existingResponses).Error; err != nil {
		tx.Rollback()
		return errors.New("error getting existing responses")
	}

	existingResponseMap := make(map[uint]models.EventFormFieldResponse)
	for _, response := range existingResponses {
		existingResponseMap[response.EventFormFieldID] = response
	}

	for _, response := range responses {
		existingResponse, exists := existingResponseMap[response.EventFormFieldID]

		if exists {
			// Update the existing response
			existingResponse.Value = *response.Value
			// Print the type of the response
			if err := tx.Model(models.EventFormFieldResponse{}).Where("id = ?", existingResponse.ID).UpdateColumn("value", response.Value).Error; err != nil {
				tx.Rollback()
				return errors.New("error updating response")
			}
			// Remove the response from the map
			delete(existingResponseMap, response.EventFormFieldID)
		} else {
			var eventFormField models.EventFormField
			if err := tx.Where("id = ?", response.EventFormFieldID).First(&eventFormField).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// Handle not found error
					return fmt.Errorf("eventFormField with ID %d not found", response.EventFormFieldID)
				}
				// Handle other DB errors
				return err
			}

			// Create a new response
			newResponse := models.EventFormFieldResponse{
				TicketID:         ticket.ID,
				EventFormFieldID: response.EventFormFieldID,
				Value:            *response.Value,
			}

			if err := tx.Create(&newResponse).Error; err != nil {
				tx.Rollback()
				return errors.New("error creating response")
			}
		}
	}

	// Delete the responses that are not included in the request
	for _, response := range existingResponseMap {
		if err := tx.Delete(&response).Error; err != nil {
			tx.Rollback()
			return errors.New("error deleting response")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
