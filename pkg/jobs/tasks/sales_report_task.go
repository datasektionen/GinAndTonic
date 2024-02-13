package tasks

// Define task types.
const (
	SalesReportType = "sales_report:generate"
)

// Define task payloads.
type SalesReportPayload struct {
	EventID int
}
