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
	EventFormFieldTypeJSON     EventFormFieldType = "json"
)

type EventFormField struct {
	gorm.Model
	EventID uint               `json:"event_id"`
	Name    string             `json:"name"`
	Type    EventFormFieldType `json:"type"`
}

// Validate validates the EventFormField model
func (field *EventFormField) Validate() error {
	switch field.Type {
	case EventFormFieldTypeText, EventFormFieldTypeCheckbox:
		return nil
	default:
		return fmt.Errorf("invalid event form field type: %s", field.Type)
	}
}
