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

	var fields []models.EventFormField = req.FormFields

	tx := effc.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var event models.Event
	// Update the form field description
	if req.FormFieldDescription != nil {
		if err := tx.Model(&event).Where("id = ?", eventID).Update("form_field_description", req.FormFieldDescription).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Delete all existing fields for the event
	if err := tx.Unscoped().Where("event_id = ?", eventID).Delete(&models.EventFormField{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Validate and insert the new fields
	for _, field := range fields {
		if err := field.Validate(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := tx.Create(&field).Error; err != nil {
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
