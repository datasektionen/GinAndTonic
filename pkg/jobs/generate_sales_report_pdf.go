package jobs

import (
	"errors"
	"fmt"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

type SaleRecord struct {
	ID          int
	CreatedAt   string // Simplified for example; you might use time.Time in a real app
	UpdatedAt   string
	DeletedAt   string
	EventID     int
	TotalSales  float64
	TicketsSold int
	Status      string
	Message     string
}

func GenerateSalesReportPDF(db *gorm.DB, data *SaleRecord, ticketReleases []models.TicketRelease) error {

	marginX := 20.0
	marginY := 20.0
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.AddPage()

	// pdf.ImageOptions("assets/logo.png", 0, 0, 65, 25, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")

	pdf.SetFont("Arial", "B", 16)
	_, lineHeight := pdf.GetFontSize()
	currentY := pdf.GetY() + lineHeight
	pdf.SetY(currentY)
	pdf.Cell(40, 10, "Konglig Datasektionen")

	pdf.Ln(-1)

	// Table data
	lineHt := 4.0

	// Set body
	pdf.SetFont("Arial", "", 10)

	// Define a smaller line height
	currentY = 50.0 // Start Y position

	// Add each field of salesData as a separate row with smaller line spacing
	pdf.Text(20, currentY, fmt.Sprintf("ID: %d", data.ID))
	currentY += lineHt
	pdf.Text(20, currentY, fmt.Sprintf("Created At: %s", data.CreatedAt))
	currentY += lineHt
	pdf.Text(20, currentY, fmt.Sprintf("Updated At: %s", data.UpdatedAt))
	currentY += lineHt
	pdf.Text(20, currentY, fmt.Sprintf("Event ID: %d", data.EventID))
	currentY += lineHt
	pdf.Text(20, currentY, fmt.Sprintf("Message: %s", data.Message))

	// Add a line break
	currentY += lineHt * 3
	pdf.SetFont("Arial", "B", 10)

	// Group ticket by ticket_request.ticket_release_id
	var subtotal float64
	var numTickets int
	// Group ticket by ticket_request.ticket_release_id
	for _, tr := range ticketReleases {
		tickets, err := models.GetAllTicketsToTicketRelease(db, tr.ID)
		if err != nil {
			return err
		}

		currency := "SEK" // TODO maybe allow multiple currencies

		for _, t := range tickets {
			var transaction models.Transaction
			if err := db.Where("ticket_id = ?", t.ID).First(&transaction).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}

			if t.IsPaid && transaction.ID != 0 {
				subtotal += float64(transaction.Amount)
				numTickets++
			}
		}

		// Check if the current Y position exceeds a certain threshold
		if currentY > 250 {
			pdf.AddPage()
			currentY = 20.0 // Reset Y position
		}

		pdf.SetFont("Arial", "", 10)

		pdf.Text(20, currentY, fmt.Sprintf("Ticket Release: %s", tr.Name))
		currentY += lineHt
		pdf.Text(20, currentY, fmt.Sprintf("Tickets sold: %d", numTickets))
		currentY += lineHt

		pdf.SetFont("Arial", "B", 10)
		pdf.Text(20, currentY, fmt.Sprintf("Subtotal: %.2f %s", subtotal/100, currency))
		currentY += lineHt * 2

		subtotal = 0.0
		numTickets = 0
	}

	// Add horizontal line
	pdf.Line(20, currentY, 190, currentY)
	currentY += lineHt

	pdf.Text(20, currentY, fmt.Sprintf("Tickets Sold: %d", data.TicketsSold))
	currentY += lineHt
	pdf.Text(20, currentY, fmt.Sprintf("Total Sales: %.2f", data.TotalSales))

	filePath := fmt.Sprintf("economy/sales_report-%d.pdf", data.ID)
	if _, err := os.Stat(filePath); err == nil {
		// File exists
		return errors.New("file already exists")
	}

	println("Saving PDF to: ", filePath)

	err := pdf.OutputFileAndClose(filePath)

	if err != nil {
		return err
	}

	return nil
}
