package models

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type EventLandingPage struct {
	gorm.Model
	EventID     uint            `json:"event_id"`
	HTML        string          `gorm:"type:text" json:"html"`
	CSS         string          `gorm:"type:text" json:"css"`
	JS          string          `gorm:"type:text" json:"js"`
	Enabled     bool            `json:"enabled" default:"false"` // Defines if the landing page is enabled, show default landing page if not enabled
	EditorState json.RawMessage `gorm:"type:json" json:"editor_state"`
}

func (elp *EventLandingPage) Validate() error {
	if elp.EventID == 0 {
		return errors.New("event_id is required")
	}

	if elp.HTML == "" {
		return errors.New("html is required")
	}

	return nil
}
