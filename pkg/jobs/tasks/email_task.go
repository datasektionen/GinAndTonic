package tasks

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

// Define task types.
const (
	TypeEmail         = "email:send"
	TypeSendOutEmail  = "email:send_out"
	TypeReminderEmail = "email:reminder"
)

// Define task payloads.
type EmailPayload struct {
	User    *models.User
	Subject string
	Content string
	EventID *uint
}

type SendOutEmailPayload struct {
	User    *models.User
	SendOut *models.SendOut
	Content string
}

type EmailReminderPayload struct {
	User       *models.User
	Subject    string
	Content    string
	ReminderID uint
}
