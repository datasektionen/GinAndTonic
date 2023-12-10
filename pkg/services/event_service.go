package services

import (
	"errors"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

type EventService struct {
	DB *gorm.DB
}

func NewEventService(db *gorm.DB) *EventService {
	return &EventService{DB: db}
}

func (es *EventService) CreateEvent(data types.EventFullWorkflowRequest, createdBy string) error {
	// Start a transaction
	tx := es.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Create Event
	event := models.Event{
		Name:           data.Event.Name,
		Description:    data.Event.Description,
		Date:           time.Unix(data.Event.Date, 0),
		Location:       data.Event.Location,
		OrganizationID: data.Event.OrganizationID,
		IsPrivate:      data.Event.IsPrivate,
		CreatedBy:      createdBy,
	}

	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		return err
	}

	println(data.TicketRelease.Name)
	println(data.TicketRelease.TicketReleaseMethodID)

	var ticketReleaseMethod models.TicketReleaseMethod
	if err := tx.First(&ticketReleaseMethod, "id = ?", data.TicketRelease.TicketReleaseMethodID).Error; err != nil {
		tx.Rollback()
		return errors.New("Invalid ticket release method ID")
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
		return errors.New("Could not create ticket release method details")
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

	var promoCode string = ""
	if data.TicketRelease.IsReserved {
		promoCode, err = utils.HashString(data.TicketRelease.PromoCode)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Create TicketRelease
	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Name:                        data.TicketRelease.Name,
		Description:                 data.TicketRelease.Description,
		Open:                        data.TicketRelease.Open,
		Close:                       data.TicketRelease.Close,
		HasAllocatedTickets:         false,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetails.ID,
		IsReserved:                  data.TicketRelease.IsReserved,
		PromoCode:                   promoCode,
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
			QuantityTotal:   uint(tt.QuantityTotal),
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
