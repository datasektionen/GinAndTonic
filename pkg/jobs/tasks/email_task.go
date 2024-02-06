package tasks

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

// Define task types.
const (
	TypeEmail         = "email:send"
	TypeReminderEmail = "email:reminder"
)

// Define task payloads.
type EmailPayload struct {
	User    *models.User
	Subject string
	Content string
	EventID *uint
}

type EmailReminderPayload struct {
	User       *models.User
	Subject    string
	Content    string
	ReminderID uint
}
