package models

import (
	"fmt"

	"gorm.io/gorm"
)

type SalesReportStatus string

const (
	SalesReportStatusPending   SalesReportStatus = "pending"
	SalesReportStatusCompleted SalesReportStatus = "completed"
	SalesReportStatusFailed    SalesReportStatus = "failed"
)

type EventSalesReport struct {
	gorm.Model
	EventID      int               `json:"event_id"`
	TotalSales   float64           `json:"total_sales"`
	TicketsSold  int               `json:"tickets_sold"`
	Status       SalesReportStatus `json:"status"`
	Message      *string           `gorm:"type:text" json:"message"`
	Transactions []Transaction     `gorm:"many2many:event_sales_report_transactions;" json:"transactions"`
	FileName     string            `json:"file_name"`
	URL          string            `gorm:"-" json:"url"` // This field will not be stored in the database
}

// Validate
func (report *EventSalesReport) Validate() error {
	switch report.Status {
	case SalesReportStatusPending, SalesReportStatusCompleted, SalesReportStatusFailed:
		return nil
	default:
		err := fmt.Errorf("invalid sales report status: %s", report.Status)
		return err
	}
}
