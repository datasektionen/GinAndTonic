package types

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
	PayWithin         string
	Tickets           []EmailTicket
}

// Associated with ticket_reserve_created
type EmalTicketReserveCreated struct {
	FullName          string
	ReserveNumber     string
	EventName         string
	TicketURL         string
	OrganizationEmail string
}

// Associated with ticket_request_confirmation
type EmailTicketRequestConfirmation struct {
	FullName          string
	EventName         string
	Tickets           []EmailTicket
	TicketURL         string
	OrganizationEmail string
}

// Associated with ticket_payment_confirmation
type EmailTicketPaymentConfirmation struct {
	FullName          string
	EventName         string
	Tickets           []EmailTicket
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
