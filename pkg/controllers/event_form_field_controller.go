package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EventFormFieldController struct {
	db *gorm.DB
}

func NewEventFormFieldController(db *gorm.DB) *EventFormFieldController {
	return &EventFormFieldController{db: db}
}

type EventFormFieldCreateRequest struct {
	FormFieldDescription *string                 `json:"form_field_description"`
	FormFields           []models.EventFormField `json:"form_fields"`
}

func (effc *EventFormFieldController) Upsert(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req EventFormFieldCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := effc.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update Event's form field description if provided
	if req.FormFieldDescription != nil {
		if err := tx.Model(&models.Event{}).Where("id = ?", eventID).Update("form_field_description", req.FormFieldDescription).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Fetch existing form fields for the event
	var existingFields []models.EventFormField
	if err := tx.Where("event_id = ?", eventID).Find(&existingFields).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	submittedFieldsMap := make(map[string]models.EventFormField)
	for _, field := range req.FormFields {
		submittedFieldsMap[field.Name] = field
	}

	existingFieldsMap := make(map[string]models.EventFormField)
	for _, field := range existingFields {
		existingFieldsMap[field.Name] = field
	}

	deletedFields := make(map[string]bool)
	for _, field := range existingFields {
		if _, exists := submittedFieldsMap[field.Name]; !exists {
			deletedFields[field.Name] = true
		}
	}

	// Upsert submitted fields
	for _, field := range req.FormFields {
		if err := field.Validate(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		field.EventID = uint(eventID)

		if field.ID == 0 {
			if _, exists := existingFieldsMap[field.Name]; exists {
				// Existing field; update
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "A field with this name already exists"})
				return
			} else {
				if err := tx.Create(&field).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}

		} else {
			// If the field name is already in the existingFieldsMap, it is an existing field
			var existingField models.EventFormField
			if err := tx.Where("event_id = ? AND name = ?", eventID, field.Name).First(&existingField).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}

			if existingField.ID != 0 && existingField.ID != field.ID {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "A field with this name already exists"})
				return
			}

			// Existing field; update
			if err := tx.Model(&models.EventFormField{}).Where("id = ?", field.ID).Updates(&field).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// Handle deletion or inactivation of fields not submitted
	for name := range deletedFields {
		if err := tx.Unscoped().Where("event_id = ? AND name = ?", eventID, name).Delete(&models.EventFormField{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Form fields upserted successfully"})
}
