package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTicketReleaseMethod(methodName, description string) *models.TicketReleaseMethod {
	return &models.TicketReleaseMethod{
		MethodName:  methodName,
		Description: description,
	}
}
