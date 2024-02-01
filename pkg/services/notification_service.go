package services

import (
	"bytes"
	"fmt"
	"html/template" // Use this for HTML templates
	"math"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"gorm.io/gorm"
)

// GenerateEmailTable generates an HTML table for the email
func GenerateEmailTable(tickets []types.EmailTicket) (string, error) {
	var emailTemplate string = `<ul class="cost-summary">`

	for _, ticket := range tickets {
		emailTemplate += fmt.Sprintf("<li><span style=\"margin: 0 10px;\">%s</span><span style=\"margin: 0 10px;\">%s SEK</span></li>", ticket.Name, ticket.Price)
	}

	emailTemplate += "</ul>"

	return emailTemplate, nil
}

// ParseTemplate parses a template file and returns the HTML content
func ParseTemplate(templateFile string, data interface{}) (string, error) {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	htmlContent := buf.String()
	// The replacement of "\x00" might not be necessary with "html/template"
	// htmlContent = strings.ReplaceAll(htmlContent, "\x00", "")

	return htmlContent, nil
}

func AddEmailJob(db *gorm.DB, user *models.User, subject, htmlContent string) {
	jobs.AddEmailJobToQueue(db, user, subject, htmlContent, nil)
}

func Notify_TicketRequestCancelled(db *gorm.DB, user *models.User, organization *models.Organization, eventName string) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	data := types.EmailTicketRequestCancelledConfirmation{
		FullName:          user.FullName(),
		EventName:         eventName,
		OrganizationEmail: organization.Email,
	}

	htmlContent, err := ParseTemplate("templates/emails/ticket_request_cancelled_confirmation.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, user, "Ticket Request Cancelled", htmlContent)

	return nil
}

func Notify_TicketCancelled(db *gorm.DB, user *models.User, organization *models.Organization, eventName string) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	data := types.EmailTicketCancelledConfirmation{
		FullName:          user.FullName(),
		EventName:         eventName,
		OrganizationEmail: organization.Email,
	}

	htmlContent, err := ParseTemplate("templates/emails/ticket_cancelled_confirmation.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, user, "Ticket Cancelled", htmlContent)

	return nil
}

func Notify_TicketAllocationCreated(db *gorm.DB, ticketId, payWithin int) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var ticket models.Ticket
	err := db.
		Preload("TicketRequest.User").
		Preload("TicketRequest.TicketRelease.Event.Organization").First(&ticket, ticketId).Error
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

	if payWithin != 0 {
		payBeforeString = utils.ConvertPayWithinToString(payWithin, ticket.UpdatedAt)
	}

	data := types.EmailTicketAllocationCreated{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		OrganizationName:  event.Organization.Name,
		OrganizationEmail: event.Organization.Email,
		PayBefore:         payBeforeString,
	}

	htmlContent, err := ParseTemplate("templates/emails/ticket_allocation_created.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, &user, fmt.Sprintf("Your ticket to %s!", event.Name), htmlContent)

	return nil
}

// Notify_TicketRequestCreated notifies the user that their ticket request has been created
func Notify_TicketRequestCreated(db *gorm.DB, ticketRequestIds []int) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var ticketRequests []models.TicketRequest
	err := db.
		Preload("User").
		Preload("TicketRelease.Event.Organization").
		Preload("TicketType").
		Where("id IN ?", ticketRequestIds).
		Find(&ticketRequests).Error
	if err != nil {
		return err
	}

	user := ticketRequests[0].User
	ticketRelease := ticketRequests[0].TicketRelease
	event := ticketRelease.Event

	if user.Email == "" {
		return fmt.Errorf("user email is empty")
	}

	var tickets []types.EmailTicket
	for _, ticket := range ticketRequests {
		tickets = append(tickets, types.EmailTicket{
			Name:  ticket.TicketType.Name,
			Price: fmt.Sprintf("%f", math.Round(100*ticket.TicketType.Price)/100),
		})
	}

	emailTicketString, _ := GenerateEmailTable(tickets)
	// emailTicketString := "<h1>Test</h1>"

	data := types.EmailTicketRequestConfirmation{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketsHTML:       template.HTML(emailTicketString), // Convert string to template.HTML
		TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/ticket-requests",
		OrganizationEmail: event.Organization.Email,
	}

	htmlContent, err := ParseTemplate("templates/emails/ticket_request_created_confirmation.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, &user, fmt.Sprintf("Your ticket request to %s!", event.Name), htmlContent)

	return nil
}

// Notify_TicketReserveCreated notifies the user that their ticket reserve has been created
func Notify_TicketPaymentConfirmation(db *gorm.DB, ticketId int) error {
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

	var tickets []types.EmailTicket
	tickets = append(tickets, types.EmailTicket{
		Name:  ticket.TicketRequest.TicketType.Name,
		Price: fmt.Sprintf("%f", math.Round(100*ticket.TicketRequest.TicketType.Price)/100),
	})

	emailTicketString, _ := GenerateEmailTable(tickets)

	data := types.EmailTicketPaymentConfirmation{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketsHTML:       template.HTML(emailTicketString),
		OrganizationEmail: event.Organization.Email,
	}

	htmlContent, err := ParseTemplate("templates/emails/ticket_payment_confirmation.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, &user, fmt.Sprintf("Ticket payment confirmation to %s!", event.Name), htmlContent)

	return nil
}
