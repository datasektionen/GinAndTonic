package services

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

type SendOutService struct {
	DB *gorm.DB
}

func NewSendOutService(db *gorm.DB) *SendOutService {
	return &SendOutService{DB: db}
}

func (sos *SendOutService) SendOutEmails(event *models.Event,
	subject string,
	message string,
	ticketReleases []models.TicketRelease,
	filters types.TicketFilter) *types.ErrorResponse {
	var allTicketRequests []models.TicketRequest
	for _, ticketRelease := range ticketReleases {
		var tr []models.TicketRequest
		if err := sos.DB.
			Preload("TicketRelease").
			Preload("User").
			Preload("Tickets").
			Where("ticket_release_id = ?", ticketRelease.ID).
			Find(&tr).Error; err != nil {
			return &types.ErrorResponse{StatusCode: 500, Message: "Error fetching tickets"}
		}
		allTicketRequests = append(allTicketRequests, tr...)
	}

	users := calculateUsers(allTicketRequests, ticketReleases, filters)

	for _, user := range users {
		err := Notify_EventSendOut(sos.DB, &user, event.Organization.Name, subject, message)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return nil
}

func applyFiltersToTickets(ticketsRequests []models.TicketRequest, filters types.TicketFilter) []models.TicketRequest {
	filteredTicketsRequests := make([]models.TicketRequest, 0)

	for _, ticketRequest := range ticketsRequests {
		keep := true

		for _, ticket := range ticketRequest.Tickets {
			checkedInMatch := filters.CheckedIn == types.Ignore || (filters.CheckedIn == types.YES && ticket.CheckedIn) || (filters.CheckedIn == types.NO && !ticket.CheckedIn)
			isPaidMatch := filters.IsPaid == types.Ignore || (filters.IsPaid == types.YES && ticket.IsPaid) || (filters.IsPaid == types.NO && !ticket.IsPaid)
			refundedMatch := filters.Refunded == types.Ignore || (filters.Refunded == types.YES && ticket.Refunded) || (filters.Refunded == types.NO && !ticket.Refunded)
			isReserveMatch := filters.IsReserve == types.Ignore || (filters.IsReserve == types.YES && ticket.IsReserve) || (filters.IsReserve == types.NO && !ticket.IsReserve)

			keep = checkedInMatch && isPaidMatch && refundedMatch && isReserveMatch

			if !keep {
				break
			}
		}

		isHandledMatch := filters.IsHandled == types.Ignore || (filters.IsHandled == types.YES && ticketRequest.IsHandled) || (filters.IsHandled == types.NO && !ticketRequest.IsHandled)
		keep = keep && isHandledMatch

		if keep {
			filteredTicketsRequests = append(filteredTicketsRequests, ticketRequest)
		}
	}

	return filteredTicketsRequests
}

func calculateUsers(tickets []models.TicketRequest, selectedTicketReleases []models.TicketRelease, filters types.TicketFilter) []models.User {
	usersMap := make(map[string]models.User)
	filteredTickets := applyFiltersToTickets(tickets, filters)

	for _, ticket := range filteredTickets {
		usersMap[ticket.UserUGKthID] = ticket.User
	}

	users := make([]models.User, 0, len(usersMap))
	for _, user := range usersMap {
		users = append(users, user)
	}
	return users
}
