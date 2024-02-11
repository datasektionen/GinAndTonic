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
	Date           int64  `json:"date"`
	Location       string `json:"location"`
	OrganizationID int    `json:"organization_id"`
	IsPrivate      bool   `json:"is_private"`
}

type EventFullWorkflowRequest struct {
	Event         EventRequest         `json:"event"`
	TicketRelease TicketReleasePostReq `json:"ticket_release"`
	TicketTypes   []TicketTypePostReq  `json:"ticket_types"`
}

type TicketReleaseFullWorkFlowRequest struct {
	TicketRelease TicketReleasePostReq `json:"ticket_release"`
	TicketTypes   []TicketTypePostReq  `json:"ticket_types"`
}

type TicketReleasePostReq struct {
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Open                  int64  `json:"open"`
	Close                 int64  `json:"close"`
	AllowExternal         bool   `json:"allow_external"`
	OpenWindowDuration    int    `json:"open_window_duration,omitempty"`
	MaxTicketsPerUser     int    `json:"max_tickets_per_user"`
	NotificationMethod    string `json:"notification_method"`
	CancellationPolicy    string `json:"cancellation_policy"`
	TicketReleaseMethodID int    `json:"ticket_release_method_id"`
	IsReserved            bool   `json:"is_reserved"`
	PromoCode             string `json:"promo_code"`
	TicketsAvailable      int    `json:"tickets_available"`
}

type TicketTypePostReq struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type ErrorResponse struct {
	StatusCode int    // HTTP status code
	Message    string // Error message
}

// Error implements error.
func (*ErrorResponse) Error() string {
	panic("unimplemented")
}

type CompleteEventWorkflowRequest struct {
}

type TicketFilter struct {
	CheckedIn bool `json:"checked_in"`
	IsHandled bool `json:"is_handled"`
	IsPaid    bool `json:"is_paid"`
	IsReserve bool `json:"is_reserve"`
	Refunded  bool `json:"refunded"`
}

type SendOutRequest struct {
	Subject          string       `json:"subject"`
	Message          string       `json:"message"`
	TicketReleaseIDs []int        `json:"ticket_release_ids"`
	Filters          TicketFilter `json:"filters"`
}
