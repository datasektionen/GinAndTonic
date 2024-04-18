package jobs

import (
	"fmt"
	"html/template"
	"math"
	"os"
	"time"

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
		Preload("TicketRequest.TicketRelease.Event.Team").
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
	if ticket.PaymentDeadline != nil {
		payBeforeString = ticket.PaymentDeadline.Format("2006-01-02 15:04:05")
	}

	data := types.EmailTicketAllocationCreated{
		FullName:  user.FullName(),
		EventName: event.Name,
		TicketURL: os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		TeamName:  event.Team.Name,
		TeamEmail: event.Team.Email,
		PayBefore: payBeforeString,
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
		FullName:    user.FullName(),
		EventName:   event.Name,
		TicketsHTML: template.HTML(emailTicketString),
		TeamEmail:   event.Team.Email,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_not_paid_in_time.html", data)
	if err != nil {
		return err
	}

	AddEmailJobToQueue(db, &user, fmt.Sprintf("Your ticket was not paid in time to %s!", event.Name), htmlContent, nil)
	return nil
}

func Notify_UpdateReserveNumbers(db *gorm.DB, ticketId int) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var ticket models.Ticket
	err := db.
		Preload("TicketRequest.TicketRelease.Event.Team").
		Preload("TicketRequest.User").
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

	var newReserveNumber int = int(ticket.ReserveNumber)

	data := types.EmailReserveUpdateNumber{
		FullName:      user.FullName(),
		EventName:     event.Name,
		TicketURL:     os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		TeamName:      event.Team.Name,
		TeamEmail:     event.Team.Email,
		ReserveNumber: fmt.Sprintf("%d", newReserveNumber),
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/reserve_update_number.html", data)

	if err != nil {
		return err
	}

	AddEmailJobToQueue(db, &user, fmt.Sprintf("Your current reserve number to %s", event.Name), htmlContent, nil)

	return nil
}

func Notify_GDPRFoodPreferencesRenewal(db *gorm.DB, user *models.User) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	if user.Email == "" {
		return fmt.Errorf("user email is empty")
	}

	data := types.EmailGDPRFoodPreferencesRenewal{
		FullName:   user.FullName(),
		ProfileURL: os.Getenv("FRONTEND_BASE_URL") + "/profile",
		RenewalURL: os.Getenv("FRONTEND_BASE_URL") + "/profile/food-preferences/renewal",
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/gdpr_food_preferences_renewal.html", data)
	if err != nil {
		return err
	}

	AddEmailJobToQueue(db, user, "Renew your food preferences consent", htmlContent, nil)

	return nil
}

func Notify_ReservedTicketAllocated(db *gorm.DB, ticketId int, paymentDeadline *time.Time) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var ticket models.Ticket
	err := db.
		Preload("TicketRequest.User").
		Preload("TicketRequest.TicketRelease.Event.Team").First(&ticket, ticketId).Error
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
	if paymentDeadline != nil {
		payBeforeString = paymentDeadline.Format("2006-01-02 15:04:05")
	}

	data := types.EmailTicketAllocationCreated{
		FullName:  user.FullName(),
		EventName: event.Name,
		TicketURL: os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		TeamName:  event.Team.Name,
		TeamEmail: event.Team.Email,
		PayBefore: payBeforeString,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_allocation_created.html", data)
	if err != nil {
		return err
	}

	AddEmailJobToQueue(db, &user, fmt.Sprintf("Your ticket to %s!", event.Name), htmlContent, nil)

	return nil
}
