package jobs

import (
	"fmt"
	"html/template"
	"math"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

func Notify_ReserveTicketConvertedAllocation(db *gorm.DB, ticketId int) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var ticket models.Ticket
	err := db.
		Preload("TicketRequest.User").
		Preload("TicketRequest.TicketRelease.Event.Organization").
		Preload("TicketRequest.TicketType").
		First(&ticket, ticketId).Error
	if err != nil {
		return err
	}

	user := ticket.TicketRequest.User
	ticketRelease := ticket.TicketRequest.TicketRelease
	event := ticketRelease.Event

	if user.Email == "" {
		return fmt.Errorf("user email is empty")
	}

	var payBeforeString string
	if ticketRelease.PayWithin != nil {
		payBeforeString = utils.ConvertPayWithinToString(int(*ticketRelease.PayWithin), ticket.UpdatedAt)
	}

	data := types.EmailTicketAllocationCreated{
		FullName:          user.FullName(),
		EventName:         event.Name,
		OrganizationName:  event.Organization.Name,
		OrganizationEmail: event.Organization.Email,
		PayBefore:         payBeforeString,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/reserve_ticket_converted_allocation.html", data)
	if err != nil {
		return err
	}

	AddEmailJobToQueue(db, &user, fmt.Sprintf("Say \"bye bye\" reserve ticket to %s!", event.Name), htmlContent, nil)

	return nil
}

// Notify_TicketReserveCreated notifies the user that their ticket reserve has been created
// We need ticket since its already been deleted from the database
func Notify_TicketNotPaidInTime(db *gorm.DB, ticket *models.Ticket) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	user := ticket.TicketRequest.User
	ticketRelease := ticket.TicketRequest.TicketRelease
	event := ticketRelease.Event

	if user.Email == "" {
		return fmt.Errorf("user email is empty")
	}

	var tickets []types.EmailTicket
	tickets = append(tickets, types.EmailTicket{
		Name:  ticket.TicketRequest.TicketType.Name,
		Price: fmt.Sprintf("%f", math.Round(100*ticket.TicketRequest.TicketType.Price)/100),
	})

	emailTicketString, _ := utils.GenerateEmailTable(tickets)

	data := types.EmailTicketNotPaidInTime{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketsHTML:       template.HTML(emailTicketString),
		OrganizationEmail: event.Organization.Email,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_not_paid_in_time.html", data)
	if err != nil {
		return err
	}

	AddEmailJobToQueue(db, &user, fmt.Sprintf("Your ticket was not paid in time to %s!", event.Name), htmlContent, nil)
	return nil
}
