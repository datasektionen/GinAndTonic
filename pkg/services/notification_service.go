package services

import (
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

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_request_cancelled_confirmation.html", data)
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

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_cancelled_confirmation.html", data)
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

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_allocation_created.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, &user, fmt.Sprintf("Your ticket to %s!", event.Name), htmlContent)

	return nil
}

func Notify_ReserveTicketAllocationCreated(db *gorm.DB, ticketId int) error {
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

	var reserveNumberString string = fmt.Sprintf("%d", ticket.ReserveNumber)

	data := types.EmailTicketAllocationReserveCreated{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		OrganizationName:  event.Organization.Name,
		OrganizationEmail: event.Organization.Email,
		ReserveNumber:     reserveNumberString,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_allocation_reserve_created.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, &user, fmt.Sprintf("Your reserve ticket to %s", event.Name), htmlContent)

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

	emailTicketString, _ := utils.GenerateEmailTable(tickets)
	// emailTicketString := "<h1>Test</h1>"

	data := types.EmailTicketRequestConfirmation{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketsHTML:       template.HTML(emailTicketString), // Convert string to template.HTML
		TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/ticket-requests",
		OrganizationEmail: event.Organization.Email,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_request_created_confirmation.html", data)
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

	emailTicketString, _ := utils.GenerateEmailTable(tickets)

	data := types.EmailTicketPaymentConfirmation{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketsHTML:       template.HTML(emailTicketString),
		OrganizationEmail: event.Organization.Email,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_payment_confirmation.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, &user, fmt.Sprintf("Ticket payment confirmation to %s!", event.Name), htmlContent)

	return nil
}

// Notify_Welcome notifies the user that they have been registered
func Notify_Welcome(db *gorm.DB, user *models.User) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	data := types.EmailWelcome{
		FullName: user.FullName(),
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/welcome_to_tessera.html", data)

	if err != nil {
		return err
	}

	AddEmailJob(db, user, "Welcome to Tessera!", htmlContent)

	return nil
}

// Notify_ExternalUserSignupVerification notifies the user that they have been registered
func Notify_ExternalUserSignupVerification(db *gorm.DB, user *models.User) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var verificationURL string = os.Getenv("FRONTEND_BASE_URL") + "/verify-email/" + user.EmailVerificationToken

	data := types.EmailExternalUserSignupVerification{
		FullName:         user.FullName(),
		VerificationLink: verificationURL,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/external_user_signup_verification.html", data)

	if err != nil {
		return err
	}

	AddEmailJob(db, user, "Verify your email", htmlContent)

	return nil
}
