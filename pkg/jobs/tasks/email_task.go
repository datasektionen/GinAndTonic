package tasks

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

// Define task types.
const (
	TypeEmail = "email:send"
)

// Define task payloads.
type EmailPayload struct {
	User    *models.User
	Subject string
	Content string
	EventID *uint
}
