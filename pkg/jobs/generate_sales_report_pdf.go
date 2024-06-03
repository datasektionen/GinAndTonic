package jobs

import (
	"errors"
	"fmt"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services/aws_service"
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
	FileName    string
	AddonsSales []models.AddOnRecord
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
	for _, tr := range ticketReleases {
		tickets, err := models.GetAllTicketsToTicketRelease(db, tr.ID)
		if err != nil {
			return err
		}

		currency := "SEK" // TODO maybe allow multiple currencies

		// We want to group tickets based on their name
		// Define a new struct that holds both the Tickets slice and the Subtotal
		type TicketGroup struct {
			Tickets  []models.Ticket
			Subtotal float64
			NumSold  int
		}

		// Then, define your map as:
		ticketGroups := make(map[string]TicketGroup)

		pdf.SetFont("Arial", "B", 10)

		pdf.Line(20, currentY, 190, currentY)
		currentY += lineHt

		pdf.Text(20, currentY, fmt.Sprintf("Ticket Release: %s", tr.Name))
		currentY += lineHt * 1.5

		// And in your loop:
		for _, t := range tickets {
			ticketRequest, err := t.GetTicketRequest(db)
			if err != nil {
				return err
			}

			var transaction models.Transaction
			if err := db.Where("ticket_id = ?", t.ID).First(&transaction).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Check if the ticket is free
					if ticketRequest.TicketType.Price == 0 {
						group := ticketGroups[ticketRequest.TicketType.Name]
						group.Subtotal += 0
						group.NumSold += ticketRequest.TicketAmount
						group.Tickets = append(group.Tickets, t)
						ticketGroups[ticketRequest.TicketType.Name] = group
					}
					continue
				}
				return err
			}

			if t.IsPaid && transaction.ID != 0 {
				group := ticketGroups[ticketRequest.TicketType.Name]
				group.Subtotal += float64(transaction.Amount)
				group.NumSold += ticketRequest.TicketAmount
				group.Tickets = append(group.Tickets, t)
				ticketGroups[ticketRequest.TicketType.Name] = group
			}
		}

		// Check if the current Y position exceeds a certain threshold
		if currentY > 250 {
			pdf.AddPage()
			currentY = 20.0 // Reset Y position
		}

		var subtotalAmount int = 0
		var subtotal float64 = 0.0
		for name, group := range ticketGroups {
			pdf.SetFont("Arial", "", 10)
			pdf.Text(20, currentY, fmt.Sprintf("Ticket: %s", name))
			currentY += lineHt
			// Subsubtotal of numSold and subtotal
			pdf.Text(20, currentY, fmt.Sprintf("Tickets sold: %d", group.NumSold))
			currentY += lineHt
			pdf.Text(20, currentY, fmt.Sprintf("Subsubtotal: %.2f %s", group.Subtotal/100, currency))
			currentY += lineHt * 2

			subtotalAmount += group.NumSold
			subtotal += group.Subtotal
		}

		pdf.SetFont("Arial", "B", 10)
		pdf.Text(20, currentY, fmt.Sprintf("Tickets sold: %d", subtotalAmount))
		currentY += lineHt

		pdf.Text(20, currentY, fmt.Sprintf("Subtotal: %.2f %s", subtotal/100, currency))
		currentY += lineHt * 2
	}

	// Check if the current Y position exceeds a certain threshold
	if currentY > 250 {
		pdf.AddPage()
		currentY = 20.0 // Reset Y position
	}

	// Add-ons sales section
	pdf.SetFont("Arial", "B", 10)
	pdf.Line(20, currentY, 190, currentY)
	currentY += lineHt

	pdf.Text(20, currentY, "Add-Ons Sales")
	currentY += lineHt * 1.5

	addonTotalSales := 0.0

	for _, addOnSale := range data.AddonsSales {
		// Check if the current Y position exceeds a certain threshold
		if currentY > 250 {
			pdf.AddPage()
			currentY = 20.0 // Reset Y position
		}

		pdf.SetFont("Arial", "", 10)
		if addOnSale.ContainsAlcohol {
			pdf.Text(20, currentY, fmt.Sprintf("Add-On: %s (Alcoholic)", addOnSale.Name))
		} else {
			pdf.Text(20, currentY, fmt.Sprintf("Add-On: %s", addOnSale.Name))
		}

		currentY += lineHt
		pdf.Text(20, currentY, fmt.Sprintf("Quantity Sold: %d", addOnSale.QuantitySales))
		currentY += lineHt
		pdf.Text(20, currentY, fmt.Sprintf("Total Sales: %.2f", addOnSale.TotalSales))
		currentY += lineHt * 2

		addonTotalSales += addOnSale.TotalSales
	}

	// Add horizontal line
	pdf.Line(20, currentY, 190, currentY)
	currentY += 1
	pdf.Line(20, currentY, 190, currentY)
	currentY += lineHt

	pdf.SetFont("Arial", "B", 10)

	pdf.Text(20, currentY, fmt.Sprintf("Tickets Sold: %d", data.TicketsSold))
	currentY += lineHt
	pdf.Text(20, currentY, fmt.Sprintf("Total Sales: %.2f", data.TotalSales))

	s3Client, err := aws_service.NewS3Client()
	if err != nil {
		sales_report_logger.Error("Error creating S3 client", err)
		return err
	}

	var folder string
	if os.Getenv("ENV") == "prod" {
		folder = "/tmp"
	} else {
		folder = "tmp"
	}

	filePath := fmt.Sprintf(folder+"/%s", data.FileName)
	// Save
	err = pdf.OutputFileAndClose(filePath)
	if err != nil {
		return err
	}

	err = aws_service.UploadFileToS3(s3Client, data.FileName, fmt.Sprintf("%s/%s", folder, data.FileName))

	if err != nil {
		sales_report_logger.Error("Error uploading file to S3", err)
		return err
	}

	if os.Getenv("ENV") == "prod" {
		err = os.Remove(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}
