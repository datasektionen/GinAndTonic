package models

import (
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

type EventFormFieldResponse struct {
	gorm.Model
	TicketRequestID  uint           `json:"ticket_request_id"`   // Foreign key to the TicketRequest model
	EventFormFieldID uint           `json:"event_form_field_id"` // Foreign key to the EventFormField model
	EventFormField   EventFormField `json:"event_form_field"`
	Value            string         `json:"value"` // The value of the field
}

// GetValueAsType returns the value of the field as the correct type
func (r *EventFormFieldResponse) GetValueAsType() (interface{}, error) {
	switch r.EventFormField.Type {
	case EventFormFieldTypeText:
		return r.Value, nil
	case EventFormFieldTypeNumber:
		return strconv.Atoi(r.Value)
	case EventFormFieldTypeCheckbox:
		return strconv.ParseBool(r.Value)
	default:
		return nil, errors.New("unknown field type")
	}
}

func (r *EventFormFieldResponse) SetValueFromType(value interface{}) error {
	switch r.EventFormField.Type {
	case EventFormFieldTypeText:
		r.Value, _ = value.(string)
	case EventFormFieldTypeNumber:
		number, ok := value.(int)
		if !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
		r.Value = strconv.Itoa(number)
	case EventFormFieldTypeCheckbox:
		boolean, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
		r.Value = strconv.FormatBool(boolean)
	default:
		return errors.New("unknown field type")
	}
	return nil
}
