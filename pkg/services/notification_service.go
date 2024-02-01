package services

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

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
	htmlContent = strings.ReplaceAll(htmlContent, "\x00", "")

	return htmlContent, nil
}

func AddEmailJob(db *gorm.DB, user *models.User, subject, htmlContent string) {
	jobs.AddEmailJobToQueue(db, user, subject, htmlContent, nil)
}

func Notify_TicketRequestCancelled(db *gorm.DB, user *models.User, organization *models.Organization, eventName string) error {
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

func Notify_TicketAllocationCreated(db *gorm.DB, ticketId int) error {
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

	var payWithin int64
	if ticketRelease.PayWithin != nil {
		payWithin = *ticketRelease.PayWithin
	}

	data := types.EmailTicketAllocationCreated{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		OrganizationName:  event.Organization.Name,
		OrganizationEmail: event.Organization.Email,
		PayWithin:         fmt.Sprintf("%d", payWithin),
	}

	htmlContent, err := ParseTemplate("templates/emails/ticket_allocation_created.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, &user, fmt.Sprintf("Your ticket to %s!", event.Name), htmlContent)

	return nil
}
