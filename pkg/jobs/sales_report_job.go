package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs/tasks"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
	"gorm.io/gorm"
)

var sales_report_logger = logrus.New()

func init() {
	// Load or create log file
	// Create logs directory if it doesn't exist
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	fileName := "logs/sales_report_job.log"

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		os.Create(fileName)
	}

	sales_report_logger_file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		sales_report_logger.Fatal(err)
	}

	sales_report_logger.SetFormatter(&logrus.JSONFormatter{})
	sales_report_logger.SetOutput(sales_report_logger_file)
	sales_report_logger.SetLevel(logrus.DebugLevel)

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

}

func AddSalesReportJobToQueue(eventID int) error {
	client := connectAsynqClient()
	defer client.Close()

	payload, err := json.Marshal(tasks.SalesReportPayload{EventID: eventID})
	if err != nil {
		return err
	}

	task := asynq.NewTask(tasks.SalesReportType, payload)
	info, err := client.Enqueue(task, asynq.Queue("sales_report"),
		asynq.MaxRetry(0),
		asynq.Timeout(30*time.Minute),
		asynq.Deadline(time.Now().Add(60*time.Minute)))

	if err != nil {
		return err
	}

	sales_report_logger.WithFields(logrus.Fields{
		"id":    string(info.ID),
		"queue": info.Queue,
	}).Info("Added sales report task to queue")

	return nil
}

func GetCompletedTransactionsByEvent(db *gorm.DB, eventID int) ([]models.Transaction, error) {
	var tickets []models.Ticket

	// Assuming db is your *gorm.DB connection and Ticket and Transaction models are properly set up
	err := db.Joins("JOIN transactions ON transactions.ticket_id = tickets.id").
		Where("transactions.event_id = ? AND transactions.status = ?", eventID, models.TransactionStatusCompleted).
		Preload("Transaction").
		Find(&tickets).Error

	if err != nil {
		fmt.Println("There was an error fetching the transactions: ", err)
		return nil, err
	}

	var transactions []models.Transaction

	// TODO revamp with surfboard

	return transactions, nil
}

func GetPaymentIntentsByEvent(db *gorm.DB, eventID int) ([]*stripe.PaymentIntent, []models.Transaction, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Get the transactions for the event
	transactions, err := GetCompletedTransactionsByEvent(db, eventID)
	if err != nil {
		return nil, nil, err
	}

	// Retrieve the corresponding payment intents from Stripe
	var paymentIntents []*stripe.PaymentIntent
	var errs []error
	var wg sync.WaitGroup
	wg.Add(len(transactions))

	for _, transaction := range transactions {
		go func(transaction models.Transaction) {
			defer wg.Done()

			var pi *stripe.PaymentIntent
			var err error
			maxRetries := 5

			for i := 0; i < maxRetries; i++ {
				pi, err = paymentintent.Get(transaction.PaymentIntentID, nil)
				if err != nil {
					stripeErr, ok := err.(*stripe.Error)
					if ok && stripeErr.Code == stripe.ErrorCodeRateLimit {
						// Wait for a bit before retrying
						time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
						continue
					}
					errs = append(errs, err)
					return
				}
				break
			}

			paymentIntents = append(paymentIntents, pi)
		}(transaction)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, nil, errs[0] // return the first error
	}

	return paymentIntents, transactions, nil
}

func GetFreeTicketsByEvent(db *gorm.DB, eventID int) ([]models.Ticket, error) {
	var tickets []models.Ticket

	// Inner join on ticket.ticket_request.ticket_type.price and eventid
	if err := db.Joins("JOIN ticket_requests ON ticket_requests.id = tickets.ticket_request_id").
		Joins("JOIN ticket_releases ON ticket_releases.id = ticket_requests.ticket_release_id").
		Joins("JOIN ticket_types ON ticket_types.id = ticket_requests.ticket_type_id").
		Where("ticket_types.price = 0 AND ticket_releases.event_id = ?", eventID).
		Preload("ticketOrder").
		Preload("ticketOrder.TicketType").
		Preload("TicketAddOns.AddOn").
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}

func GenerateReportForEvent(eventID int, db *gorm.DB) (*models.EventSalesReport, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	paymentIntents, transactions, err := GetPaymentIntentsByEvent(db, eventID)
	if err != nil {

		return nil, err
	}

	// We need to get all Ticket.TicketAddOns
	var purchasedTickets []models.Ticket
	for _, transaction := range transactions {
		var ticket models.Ticket
		if err := db.Preload("TicketAddOns.AddOn").Where("id = ?", transaction.TicketID).First(&ticket).Error; err != nil {
			return nil, err
		}

		purchasedTickets = append(purchasedTickets, ticket)
	}

	freeTickets, err := GetFreeTicketsByEvent(db, eventID)
	if err != nil {
		return nil, err
	}

	// Combine purchased tickets and free tickets and iterate over them
	allTickets := append(purchasedTickets, freeTickets...)
	var allSoldAddOns map[uint]*models.AddOnRecord = make(map[uint]*models.AddOnRecord)

	for _, ticket := range allTickets {
		if len(ticket.TicketAddOns) > 0 {
			for _, addOn := range ticket.TicketAddOns {
				// Create a models.AddOnRecord struct if it doesn't exist
				if _, ok := allSoldAddOns[addOn.AddOnID]; !ok {
					allSoldAddOns[addOn.AddOnID] = &models.AddOnRecord{
						ID:              addOn.AddOnID,
						Name:            addOn.AddOn.Name,
						QuantitySales:   0,
						TotalSales:      0,
						ContainsAlcohol: addOn.AddOn.ContainsAlcohol,
					}
				}

				// Increment the quantity of the AddOnRecord struct
				allSoldAddOns[addOn.AddOnID].QuantitySales += addOn.Quantity
				allSoldAddOns[addOn.AddOnID].TotalSales += addOn.AddOn.Price * float64(addOn.Quantity)
			}
		}
	}

	// Convert map to array
	var updatedAddOnsSales []models.AddOnRecord
	for _, addOn := range allSoldAddOns {
		updatedAddOnsSales = append(updatedAddOnsSales, *addOn)
	}

	randomUUID := uuid.New()
	fileName := fmt.Sprintf("sales_report-%d-%s.pdf", eventID, randomUUID.String())

	// Initialize your report structure
	report := &models.EventSalesReport{
		EventID:      eventID,
		TotalSales:   0,
		TicketsSold:  0,
		Status:       models.SalesReportStatusPending,
		Transactions: transactions,
		FileName:     fileName,
		AddOnsSales:  updatedAddOnsSales,
	}

	for _, pi := range paymentIntents {
		// Sum total sales and count tickets
		report.TotalSales += float64(pi.Amount)
		report.TicketsSold += 1 // TODO: Should be changed when we support multiple tickets per payment intent
	}

	report.TicketsSold += len(freeTickets)

	// Convert total sales from cents to a more readable format if necessary
	report.TotalSales = report.TotalSales / 100

	SaveReport(db, report)

	return report, nil
}

func SaveReport(db *gorm.DB, report *models.EventSalesReport) error {
	if err := db.Create(&report).Error; err != nil {
		return err
	}
	return nil
}

func SalesReportSetSuccess(db *gorm.DB, report *models.EventSalesReport) error {
	report.Status = models.SalesReportStatusCompleted
	if err := db.Save(&report).Error; err != nil {
		return err
	}
	return nil
}

func SalesReportSetFailed(db *gorm.DB, report *models.EventSalesReport, message string) error {
	report.Status = models.SalesReportStatusFailed
	report.Message = &message
	if err := db.Save(&report).Error; err != nil {
		return err
	}
	return nil
}

func HandleSalesReportJob(db *gorm.DB) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p tasks.SalesReportPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		report, err := GenerateReportForEvent(p.EventID, db)
		if err != nil {
			SalesReportSetFailed(db, &models.EventSalesReport{EventID: p.EventID}, err.Error())
			return err
		}

		var msg string
		if report.Message != nil {
			msg = *report.Message
		} else {
			msg = ""
		}

		salesData := SaleRecord{
			ID:          int(report.ID),
			CreatedAt:   report.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   report.UpdatedAt.Format("2006-01-02 15:04:05"),
			DeletedAt:   "",
			EventID:     report.EventID,
			TotalSales:  report.TotalSales,
			TicketsSold: report.TicketsSold,
			Status:      string(report.Status),
			Message:     msg,
			FileName:    report.FileName,
			AddonsSales: report.AddOnsSales,
		}

		trs, err := models.GetTicketReleasesToEvent(db, uint(p.EventID))
		if err != nil {
			SalesReportSetFailed(db, report, err.Error())
			return err
		}

		err = GenerateSalesReportPDF(db, &salesData, trs)

		if err != nil {
			SalesReportSetFailed(db, report, err.Error())
			return err
		}

		err = SalesReportSetSuccess(db, report)
		if err != nil {
			SalesReportSetFailed(db, report, err.Error())
			return err
		}

		return nil
	}
}
