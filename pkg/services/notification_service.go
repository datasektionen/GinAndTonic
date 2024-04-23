package services

import (
	"fmt"
	"html/template" // Use this for HTML templates
	"math"
	"os"
	"time"

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

func Notify_TicketAllocationCreated(db *gorm.DB, ticketId int, paymentDeadline *time.Time) error {
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
	if paymentDeadline != nil {
		payBeforeString = paymentDeadline.Format("2006-01-02 15:04:05")
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
			Price: fmt.Sprintf("%.2f", math.Round(100*ticket.TicketType.Price)/100)})
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
		Price: fmt.Sprintf("%.2f", math.Round(100*ticket.TicketRequest.TicketType.Price)/100),
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

	var verificationURL string = os.Getenv("FRONTEND_BASE_URL") + "/verify-email/" + *user.EmailVerificationToken

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

func Notify_RemindUserOfTicketRelease(db *gorm.DB, trReminder *models.TicketReleaseReminder) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var ticketRelease models.TicketRelease
	err := db.Preload("Event.Organization").First(&ticketRelease, trReminder.TicketReleaseID).Error
	if err != nil {
		return fmt.Errorf("ticket release not found")
	}

	var user models.User
	err = db.Where("ug_kth_id = ?", trReminder.UserUGKthID).First(&user).Error
	if err != nil {
		return fmt.Errorf("user not found")
	}

	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	data := types.EmailTicketReleaseReminder{
		FullName:          user.FullName(),
		EventName:         ticketRelease.Event.Name,
		TicketReleaseName: ticketRelease.Name,
		EventURL:          os.Getenv("FRONTEND_BASE_URL") + "/events/" + fmt.Sprintf("%d", ticketRelease.EventID),
		OpensAt:           (time.Unix(ticketRelease.Open, 0).In(loc)).Format("2006-01-02 15:04:05"),
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_release_reminder.html", data)
	if err != nil {
		return err
	}

	jobs.AddReminderEmailJobToQueueAt(db, &user,
		fmt.Sprintf("Ticket release reminder for %s", ticketRelease.Event.Name),
		htmlContent, trReminder.ID, trReminder.ReminderTime)

	return nil
}

func Notify_PasswordReset(db *gorm.DB, pwReset *models.UserPasswordReset) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var resetURL string = os.Getenv("FRONTEND_BASE_URL") + "/reset-password/" + pwReset.Token

	data := types.EmailPasswordReset{
		ResetLink: resetURL,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/password_reset.html", data)

	if err != nil {
		return err
	}

	AddEmailJob(db, &pwReset.User, "Reset your password", htmlContent)

	return nil
}

func Notify_EventSendOut(db *gorm.DB, sendOut *models.SendOut, user *models.User, message string) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	jobs.AddSendOutEmailJobToQueue(db, user, sendOut, message)

	return nil
}

func Notify_RequestChangePreferredEmail(db *gorm.DB,
	user *models.User,
	preferredEmail *models.PreferredEmail) error {
	if os.Getenv("ENV") == "test" {
		return nil
	}

	var verificationURL string = os.Getenv("FRONTEND_BASE_URL") + "/verify-preferred-email/" + preferredEmail.Token

	data := types.EmailRequestChangePreferredEmail{
		VerificationLink: verificationURL,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/request_change_preferred_email.html", data)

	if err != nil {
		return err
	}

	AddEmailJob(db, user, "New preferred email validation", htmlContent)

	return nil
}

func Notify_UpdatedPaymentDeadlineEmail(db *gorm.DB, ticketId int, paymentDeadline *time.Time) error {
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
	if paymentDeadline != nil {
		payBeforeString = paymentDeadline.Format("2006-01-02 15:04:05")
	}

	data := types.EmailUpdatePaymentDeadline{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		OrganizationEmail: event.Organization.Email,
		PayBefore:         payBeforeString,
	}

	htmlContent, err := utils.ParseTemplate("templates/emails/ticket_updated_payment_deadine.html", data)
	if err != nil {
		return err
	}

	AddEmailJob(db, &user, fmt.Sprintf("Updated payment deadline for %s", event.Name), htmlContent)

	return nil
}
