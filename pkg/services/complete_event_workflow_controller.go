package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	feature_services "github.com/DowLucas/gin-ticket-release/pkg/services/features"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

type CompleteEventWorkflowService struct {
	DB *gorm.DB
}

func NewCompleteEventWorkflowService(db *gorm.DB) *CompleteEventWorkflowService {
	return &CompleteEventWorkflowService{DB: db}
}

func GenerateReferenceID(tx *gorm.DB) (*string, error) {
	maxAttempts := 10
	var refId string
	for i := 0; i < maxAttempts; i++ {
		referenceID := utils.GenerateRandomString(10)

		var exEvent models.Event
		if err := tx.First(&exEvent, "reference_id = ?", referenceID).Error; err != nil {
			// If the error is a record not found error, the referenceID is unique
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Set the unique referenceID and break the loop
				refId = referenceID
				break
			} else {
				// If there's another kind of error, rollback and return it
				tx.Rollback()
				return nil, err
			}
		}

		// If a unique referenceID couldn't be generated after maxAttempts, rollback and return an error
		if i == maxAttempts-1 {
			tx.Rollback()
			return nil, errors.New("could not generate unique reference ID")
		}
	}

	return &refId, nil
}

func (es *CompleteEventWorkflowService) CreateEvent(data types.EventFullWorkflowRequest, user *models.User) (revent *models.Event, err error) {
	// Start a transaction
	tx := es.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return nil, tx.Error
	}

	// Check if reference ID is unique
	refId, err := GenerateReferenceID(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create Event
	event := models.Event{
		ReferenceID:    *refId,
		Name:           data.Event.Name,
		Description:    data.Event.Description,
		Date:           time.Unix(data.Event.Date, 0),
		Location:       data.Event.Location,
		OrganizationID: data.Event.OrganizationID,
		IsPrivate:      data.Event.IsPrivate,
		CreatedBy:      user.UGKthID,
	}

	if data.Event.EndDate != nil {
		ed := time.Unix(*data.Event.EndDate, 0)
		event.EndDate = &ed
	}

	token, err := utils.GenerateSecretToken()
	if err != nil {
		return nil, err
	}

	event.SecretToken = token

	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var ticketReleaseMethod models.TicketReleaseMethod
	if err := tx.First(&ticketReleaseMethod, "id = ?", data.TicketRelease.TicketReleaseMethodID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("invalid ticket release method ID")
	}

	ticketReleaseMethodDetails := models.TicketReleaseMethodDetail{
		TicketReleaseMethodID: uint(data.TicketRelease.TicketReleaseMethodID),
		OpenWindowDuration:    int64(data.TicketRelease.OpenWindowDuration),
		NotificationMethod:    data.TicketRelease.NotificationMethod,
		CancellationPolicy:    data.TicketRelease.CancellationPolicy,
		MaxTicketsPerUser:     uint(data.TicketRelease.MaxTicketsPerUser),
		MethodDescription:     data.TicketRelease.MethodDescription,
	}

	if err := ticketReleaseMethodDetails.Validate(); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(&ticketReleaseMethodDetails).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("could not create ticket release method details")
	}

	method, err := models.NewTicketReleaseConfig(ticketReleaseMethod.MethodName, &ticketReleaseMethodDetails)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := method.Validate(); err != nil {
		tx.Rollback()
		return nil, err
	}

	var promoCode string
	if data.TicketRelease.IsReserved {
		if data.TicketRelease.PromoCode == "" {
			tx.Rollback()
			return nil, errors.New("promo code is required for reserved ticket releases")
		}

		promoCode, err = utils.EncryptString(data.TicketRelease.PromoCode)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("could not hash promo code")
		}
	}

	// Create TicketRelease
	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Name:                        data.TicketRelease.Name,
		Description:                 data.TicketRelease.Description,
		Open:                        data.TicketRelease.Open,
		Close:                       data.TicketRelease.Close,
		TicketsAvailable:            data.TicketRelease.TicketsAvailable,
		HasAllocatedTickets:         false,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetails.ID,
		IsReserved:                  data.TicketRelease.IsReserved,
		PromoCode:                   &promoCode,
		AllowExternal:               data.TicketRelease.AllowExternal,
	}

	if err := tx.Create(&ticketRelease).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create TicketTypes
	for _, tt := range data.TicketTypes {
		ticketType := models.TicketType{
			EventID:         event.ID,
			Name:            tt.Name,
			Description:     tt.Description,
			Price:           tt.Price,
			TicketReleaseID: ticketRelease.ID,
		}

		if err := tx.Create(&ticketType).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	eventID := fmt.Sprint(event.ID)
	ticketReleaseID := fmt.Sprint(ticketRelease.ID)
	err = feature_services.IncrementFeatureUsage(tx, user.Network.PlanEnrollment.ID, "max_ticket_releases_per_event", &eventID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = feature_services.IncrementFeatureUsage(tx, user.Network.PlanEnrollment.ID, "max_events", nil)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = feature_services.IncrementFeatureUsage(tx, user.Network.PlanEnrollment.ID, "max_ticket_types_per_ticket_release", &ticketReleaseID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Commit the transaction
	return &event, nil
}

func (es *CompleteEventWorkflowService) CreateTicketRelease(data types.TicketReleaseFullWorkFlowRequest, eventID int, UGKthId string) error {
	// Start a transaction
	tx := es.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// Check if the user is a super admin
	var event models.Event
	if err := tx.First(&event, eventID).Error; err != nil {
		tx.Rollback()
		return errors.New("invalid event ID")
	}

	// Find ticket release method
	var ticketReleaseMethod models.TicketReleaseMethod
	if err := tx.First(&ticketReleaseMethod, "id = ?", data.TicketRelease.TicketReleaseMethodID).Error; err != nil {
		tx.Rollback()
		return errors.New("invalid ticket release method ID")
	}

	ticketReleaseMethodDetails := models.TicketReleaseMethodDetail{
		TicketReleaseMethodID: uint(data.TicketRelease.TicketReleaseMethodID),
		OpenWindowDuration:    int64(data.TicketRelease.OpenWindowDuration),
		NotificationMethod:    data.TicketRelease.NotificationMethod,
		CancellationPolicy:    data.TicketRelease.CancellationPolicy,
		MaxTicketsPerUser:     uint(data.TicketRelease.MaxTicketsPerUser),
	}

	if err := ticketReleaseMethodDetails.Validate(); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Create(&ticketReleaseMethodDetails).Error; err != nil {
		tx.Rollback()
		return errors.New("could not create ticket release method details")
	}

	method, err := models.NewTicketReleaseConfig(ticketReleaseMethod.MethodName, &ticketReleaseMethodDetails)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := method.Validate(); err != nil {
		tx.Rollback()
		return err
	}

	var promoCode string
	if data.TicketRelease.IsReserved {
		if data.TicketRelease.PromoCode == "" {
			tx.Rollback()
			return errors.New("promo code is required for reserved ticket releases")
		}

		promoCode, err = utils.EncryptString(data.TicketRelease.PromoCode)
		if err != nil {
			tx.Rollback()
			return errors.New("could not hash promo code")
		}
	}

	// Create TicketRelease
	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Name:                        data.TicketRelease.Name,
		Description:                 data.TicketRelease.Description,
		Open:                        data.TicketRelease.Open,
		Close:                       data.TicketRelease.Close,
		TicketsAvailable:            data.TicketRelease.TicketsAvailable,
		HasAllocatedTickets:         false,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetails.ID,
		IsReserved:                  data.TicketRelease.IsReserved,
		PromoCode:                   &promoCode,
		AllowExternal:               data.TicketRelease.AllowExternal,
	}

	if err := tx.Create(&ticketRelease).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create TicketTypes
	for _, tt := range data.TicketTypes {
		ticketType := models.TicketType{
			EventID:         event.ID,
			Name:            tt.Name,
			Description:     tt.Description,
			Price:           tt.Price,
			TicketReleaseID: ticketRelease.ID,
		}

		if err := tx.Create(&ticketType).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	return tx.Commit().Error
}
