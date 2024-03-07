package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

	for _, ticket := range tickets {
		if ticket.Transaction == nil {
			continue
		}
		transactions = append(transactions, *ticket.Transaction)
	}

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
	for _, transaction := range transactions {
		pi, err := paymentintent.Get(transaction.PaymentIntentID, nil)
		if err != nil {
			return nil, nil, err
		}
		paymentIntents = append(paymentIntents, pi)
	}

	return paymentIntents, transactions, nil
}

func GenerateReportForEvent(eventID int, db *gorm.DB) (*models.EventSalesReport, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	paymentIntents, transactions, err := GetPaymentIntentsByEvent(db, eventID)
	if err != nil {

		return nil, err
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
	}

	for _, pi := range paymentIntents {
		// Sum total sales and count tickets
		report.TotalSales += float64(pi.Amount)
		report.TicketsSold += 1 // TODO: Should be changed when we support multiple tickets per payment intent
	}

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
