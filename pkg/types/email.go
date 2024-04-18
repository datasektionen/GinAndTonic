package types

import "html/template"

type EmailTicket struct {
	Name  string
	Price string
}

// Associated with ticket_allocation_created
type EmailTicketAllocationCreated struct {
	FullName  string
	EventName string
	TicketURL string
	TeamName  string
	TeamEmail string
	PayBefore string
}

// Associated with ticket_reserve_created
type EmailTicketAllocationReserveCreated struct {
	FullName      string
	ReserveNumber string
	EventName     string
	TicketURL     string
	TeamName      string
	TeamEmail     string
}

// Associated with ticket_request_confirmation
type EmailTicketRequestConfirmation struct {
	FullName    string
	EventName   string
	TicketsHTML template.HTML
	TicketURL   string
	TeamEmail   string
}

// Associated with ticket_payment_confirmation
type EmailTicketPaymentConfirmation struct {
	FullName    string
	EventName   string
	TicketsHTML template.HTML
	TeamEmail   string
}

// Associated with ticket_cancelled_confirmation
type EmailTicketCancelledConfirmation struct {
	FullName  string
	EventName string
	TeamEmail string
}

// Associated with ticket_request_cancelled_confirmation
type EmailTicketRequestCancelledConfirmation struct {
	FullName  string
	EventName string
	TeamEmail string
}

// Associated with ticket_request_reserve_number_update
type EmailTicketRequestReserveNumberUpdate struct {
	FullName      string
	EventName     string
	TicketURL     string
	RequestNumber string
	TeamEmail     string
}

type EmailTicketNotPaidInTime struct {
	FullName    string
	EventName   string
	TicketsHTML template.HTML
	TeamEmail   string
}

type EmailReserveTicketConvertedAllocation struct {
	FullName  string
	EventName string
	TeamEmail string
	TeamName  string
}

type EmailReserveUpdateNumber struct {
	FullName      string
	EventName     string
	TicketURL     string
	TeamEmail     string
	TeamName      string
	ReserveNumber string
}

// EmailWelcome is the struct for the welcome email
type EmailWelcome struct {
	FullName string
}

// EmailExternalUserSignupVerification is the struct for the external user signup verification email
type EmailExternalUserSignupVerification struct {
	FullName         string
	VerificationLink string
}

// EmailContact is the struct for the contact email
type EmailContact struct {
	FullName string
	TeamName string
	Subject  string
	Message  string
	Email    string
}

type EmailTicketReleaseReminder struct {
	FullName          string
	EventName         string
	TicketReleaseName string
	EventURL          string
	OpensAt           string
}

type EmailPasswordReset struct {
	ResetLink string
}

type EmailEventSendOut struct {
	Message  template.HTML
	TeamName string
}

type EmailRequestChangePreferredEmail struct {
	VerificationLink string
}

type EmailUpdatePaymentDeadline struct {
	FullName  string
	EventName string
	TicketURL string
	PayBefore string
	TeamEmail string
}
