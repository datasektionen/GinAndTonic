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
	filteredTicketsRequests := ticketsRequests
	if filters.CheckedIn {
		filteredTicketsRequests = filterTickets(filteredTicketsRequests, func(ticketRequest models.TicketRequest) bool {
			for _, ticket := range ticketRequest.Tickets {
				if ticket.CheckedIn {
					return true
				}
			}
			return false
		})
	}
	if filters.IsPaid {
		filteredTicketsRequests = filterTickets(filteredTicketsRequests, func(ticketRequest models.TicketRequest) bool {
			for _, ticket := range ticketRequest.Tickets {
				if ticket.IsPaid {
					return true
				}
			}
			return false
		})
	}
	if filters.Refunded {
		filteredTicketsRequests = filterTickets(filteredTicketsRequests, func(ticketRequest models.TicketRequest) bool {
			for _, ticket := range ticketRequest.Tickets {
				if ticket.Refunded {
					return true
				}
			}
			return false
		})
	}
	if filters.IsReserve {
		filteredTicketsRequests = filterTickets(filteredTicketsRequests, func(ticketRequest models.TicketRequest) bool {
			for _, ticket := range ticketRequest.Tickets {
				if ticket.IsReserve {
					return true
				}
			}
			return false
		})
	}
	if filters.IsHandled {
		filteredTicketsRequests = filterTickets(filteredTicketsRequests, func(ticketRequest models.TicketRequest) bool {
			return ticketRequest.IsHandled == filters.IsHandled
		})
	}
	return filteredTicketsRequests
}

func filterTickets(tickets []models.TicketRequest, condition func(models.TicketRequest) bool) []models.TicketRequest {
	filteredTickets := []models.TicketRequest{}
	for _, ticket := range tickets {
		if condition(ticket) {
			filteredTickets = append(filteredTickets, ticket)
		}
	}
	return filteredTickets
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
