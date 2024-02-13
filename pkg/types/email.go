package types

import "html/template"

type EmailTicket struct {
	Name  string
	Price string
}

// Associated with ticket_allocation_created
type EmailTicketAllocationCreated struct {
	FullName          string
	EventName         string
	TicketURL         string
	OrganizationName  string
	OrganizationEmail string
	PayBefore         string
}

// Associated with ticket_reserve_created
type EmailTicketAllocationReserveCreated struct {
	FullName          string
	ReserveNumber     string
	EventName         string
	TicketURL         string
	OrganizationName  string
	OrganizationEmail string
}

// Associated with ticket_request_confirmation
type EmailTicketRequestConfirmation struct {
	FullName          string
	EventName         string
	TicketsHTML       template.HTML
	TicketURL         string
	OrganizationEmail string
}

// Associated with ticket_payment_confirmation
type EmailTicketPaymentConfirmation struct {
	FullName          string
	EventName         string
	TicketsHTML       template.HTML
	OrganizationEmail string
}

// Associated with ticket_cancelled_confirmation
type EmailTicketCancelledConfirmation struct {
	FullName          string
	EventName         string
	OrganizationEmail string
}

// Associated with ticket_request_cancelled_confirmation
type EmailTicketRequestCancelledConfirmation struct {
	FullName          string
	EventName         string
	OrganizationEmail string
}

// Associated with ticket_payment_reminder
type EmailTicketPaymentReminder struct {
	FullName          string
	EventName         string
	TicketURL         string
	OrganizationName  string
	OrganizationEmail string
	PayWithin         string
}

// Associated with ticket_request_reserve_number_update
type EmailTicketRequestReserveNumberUpdate struct {
	FullName          string
	EventName         string
	TicketURL         string
	RequestNumber     string
	OrganizationEmail string
}

type EmailTicketNotPaidInTime struct {
	FullName          string
	EventName         string
	TicketsHTML       template.HTML
	OrganizationEmail string
}

type EmailReserveTicketConvertedAllocation struct {
	FullName          string
	EventName         string
	OrganizationEmail string
	OrganizationName  string
}

type EmailReserveUpdateNumber struct {
	FullName          string
	EventName         string
	TicketURL         string
	OrganizationEmail string
	OrganizationName  string
	ReserveNumber     string
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
	FullName         string
	OrganizationName string
	Subject          string
	Message          string
	Email            string
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
	Message          template.HTML
	OrganizationName string
}

type EmailRequestChangePreferredEmail struct {
	VerificationLink string
}
