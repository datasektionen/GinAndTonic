package surfboard_types

type SurfboardOrder struct {
	TerminalID string               `json:"terminal$id"`
	Type       string               `json:"type"`
	OrderLines []SurfboardOrderLine `json:"orderLines"`
}

type SurfboardOrderLine struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Quantity   int                 `json:"quantity"`
	ItemAmount SurfboardItemAmount `json:"itemAmount"`
}

type SurfboardItemAmount struct {
	Regular  int            `json:"regular"`
	Total    int            `json:"total"`
	Currency string         `json:"currency"`
	Tax      []SurfboardTax `json:"tax"`
}

type SurfboardTax struct {
	Amount     int    `json:"amount"`
	Percentage int    `json:"percentage"`
	Type       string `json:"type"`
}
