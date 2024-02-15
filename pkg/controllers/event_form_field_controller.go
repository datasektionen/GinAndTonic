package controllers

import (
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

func (effc *EventFormFieldController) Upsert(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var fields []models.EventFormField
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add the event ID to the fields
	for i := range fields {
		fields[i].EventID = uint(eventID)
	}

	tx := effc.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingFields []models.EventFormField
	if err := tx.Where("event_id = ?", eventID).Find(&existingFields).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existingFieldMap := make(map[string]models.EventFormField)
	for _, field := range existingFields {
		existingFieldMap[field.Name] = field
	}

	for _, field := range fields {
		if err := field.Validate(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		existingField, exists := existingFieldMap[field.Name]
		if exists {
			// Update the existing field
			existingField.Type = field.Type
			existingField.Description = field.Description // Add this line
			if err := tx.Save(&existingField).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// Remove the field from the map
			delete(existingFieldMap, field.Name)
		} else {
			// Create a new field
			if err := tx.Create(&field).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// Delete the fields that are not included in the request
	for _, field := range existingFieldMap {
		// Delete the field
		if err := tx.Unscoped().Delete(&field).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"new_fields": fields})
}
