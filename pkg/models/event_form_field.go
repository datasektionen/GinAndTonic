package models

import (
	"fmt"

	"gorm.io/gorm"
)

type EventFormFieldType string

const (
	EventFormFieldTypeText     EventFormFieldType = "text"
	EventFormFieldTypeCheckbox EventFormFieldType = "checkbox"
	EventFormFieldTypeNumber   EventFormFieldType = "number"
)

type EventFormField struct {
	gorm.Model
	EventID     uint               `json:"event_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Type        EventFormFieldType `json:"type"`
}

// Validate validates the EventFormField model
func (field *EventFormField) Validate() error {
	switch field.Type {
	case EventFormFieldTypeText, EventFormFieldTypeCheckbox, EventFormFieldTypeNumber:
		return nil
	default:
		return fmt.Errorf("invalid event form field type: %s", field.Type)
	}
}
