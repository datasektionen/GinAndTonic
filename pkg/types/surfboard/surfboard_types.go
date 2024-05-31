package surfboard_types

type SurfboardRequestArgs struct {
	Endpoint   string
	Method     string
	MerchantId *string
	PartnerId  *string
	StoreId    *string
	JSONStr    *[]byte
}

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

type OrderWebhookType string

const (
	OrderCancelled        OrderWebhookType = "order.cancelled"
	OrderUpdated          OrderWebhookType = "order.updated"
	OrderDeleted          OrderWebhookType = "order.deleted"
	OrderTerminalEvent    OrderWebhookType = "order.terminal.event"
	OrderPaymentInit      OrderWebhookType = "order.paymentinitiated"
	OrderPaymentProc      OrderWebhookType = "order.paymentprocessed"
	OrderPaymentComp      OrderWebhookType = "order.paymentcompleted"
	OrderPaymentFailed    OrderWebhookType = "order.paymentfailed"
	OrderPaymentCancelled OrderWebhookType = "order.paymentcancelled"
	OrderPaymentVoided    OrderWebhookType = "order.paymentvoided"
)

type OrderWebhookData struct {
	EventType OrderWebhookType `json:"eventType"`
	Metadata  struct {
		EventID        string `json:"eventId"`
		Created        int    `json:"created,omitempty"`
		RetryAttempt   int    `json:"retryAttempt"`
		WebhookEventId string `json:"webhookEventId"`
	} `json:"metadata"`
	Data struct {
		OrderID                     string `json:"orderId"`
		PaymentID                   string `json:"paymentId,omitempty"`
		OrderStatus                 string `json:"orderStatus,omitempty"`
		TransactionID               string `json:"transactionId,omitempty"`
		PaymentMethod               string `json:"paymentMethod,omitempty"`
		PaymentStatus               string `json:"paymentStatus,omitempty"`
		TruncatedPan                string `json:"truncatedPan,omitempty"`
		CardLabel                   string `json:"cardLabel,omitempty"`
		PosEntryMode                string `json:"posEntryMode,omitempty"`
		IssuerApplication           string `json:"issuerApplication,omitempty"`
		TerminalVerificationResult  string `json:"terminalVerificationResult,omitempty"`
		Aid                         string `json:"aid,omitempty"`
		CustomerResponseCode        string `json:"customerResponseCode,omitempty"`
		CvmMethod                   string `json:"cvmMethod,omitempty"`
		AuthMode                    string `json:"authMode,omitempty"`
		CustomerResponseDescription string `json:"customerResponseDescription,omitempty"`
	} `json:"data,omitempty"`
}
