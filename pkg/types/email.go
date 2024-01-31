package types

type EmailTicketAllocationCreated struct {
	FullName          string
	EventName         string
	TicketURL         string
	OrganizationName  string
	OrganizationEmail string
	PayWithin         string
}

type EmalTicketReserveCreated struct {
	FullName          string
	ReserveNumber     string
	EventName         string
	TicketURL         string
	OrganizationName  string
	OrganizationEmail string
}

type EmailTicketRequestConfirmed struct {
	FullName          string
	EventName         string
	TicketAmount      string
	TicketType        string
	TicketURL         string
	OrganizationEmail string
}
