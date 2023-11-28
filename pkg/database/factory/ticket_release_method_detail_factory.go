package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTicketReleaseMethodDetail(maxTicketsPerUser uint, notificationMethod, cancellationPolicy string, openWindowDuration int64, ticketReleaseMethodID uint) *models.TicketReleaseMethodDetail {
	return &models.TicketReleaseMethodDetail{
		MaxTicketsPerUser:     maxTicketsPerUser,
		NotificationMethod:    notificationMethod,
		CancellationPolicy:    cancellationPolicy,
		OpenWindowDuration:    openWindowDuration,
		TicketReleaseMethodID: ticketReleaseMethodID,
	}
}
