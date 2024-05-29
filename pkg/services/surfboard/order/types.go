package surfboard_service_order

type OrderRequest struct {
	TerminalID       string           `json:"terminal$id"`
	Type             string           `json:"type"`
	OrderLines       []OrderLine      `json:"orderLines"`
	ControlFunctions ControlFunctions `json:"controlFunctions,omitempty"`
}

type OrderLine struct {
	ID             string     `json:"id"`
	CategoryID     string     `json:"categoryId,omitempty"`
	Description    string     `json:"description,omitempty"`
	ImageUrl       string     `json:"imageUrl,omitempty"`
	ExternalItemId string     `json:"externalItemId,omitempty"`
	Name           string     `json:"name"`
	Quantity       int        `json:"quantity"`
	ItemAmount     ItemAmount `json:"itemAmount"`
}

type ItemAmount struct {
	Regular  int    `json:"regular"`
	Total    int    `json:"total"`
	Currency string `json:"currency"`
	Tax      []Tax  `json:"tax"`
}

type Tax struct {
	Amount     int    `json:"amount"`
	Percentage int    `json:"percentage"`
	Type       string `json:"type"`
}

type ControlFunctions struct {
	CancelPreviousPendingOrder bool `json:"cancelPreviousPendingOrder,omitempty"`
}
