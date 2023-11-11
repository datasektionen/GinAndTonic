package types

type Body struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	User      string `json:"user"`
	Emails    string `json:"emails"`
	UGKthID   string `json:"ugkthid"`
}

type TicketTypesRequest struct {
	Name                    string  `json:"name"`
	Description             string  `json:"description"`
	Price                   float64 `json:"price"`
	QuantityTotal           uint    `json:"quantity_total"`
	IsReserved              bool    `json:"is_reserved"`
	TicketReleaseMethodName string  `json:"ticket_release_method_name"`
}

type TicketReleaseRequest struct {
	Open  uint `json:"open"`
	Close uint `json:"close"`
}

type EventRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Date           uint   `json:"date"`
	Location       string `json:"location"`
	OrganizationID int    `json:"organization_id"`
}
