package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs/tasks"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var notification_logger = logrus.New()

func init() {
	// Load or create log file
	notification_log_file, err := os.OpenFile("logs/notification.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		notification_logger.Fatal(err)
	}

	notification_logger.SetFormatter(&logrus.JSONFormatter{})
	notification_logger.SetOutput(notification_log_file)
	notification_logger.SetLevel(logrus.DebugLevel)
}

func CreateNotificationInDb(db *gorm.DB, notification *models.Notification) error {
	// Validate
	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create notification
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"error":        err,
		}).Error("Error creating notification in database")

		return err
	}

	return tx.Commit().Error
}

func AddEmailJobToQueue(db *gorm.DB, user *models.User, subject, content string, evnetId *uint) error {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
	defer client.Close()

	payload, err := json.Marshal(tasks.EmailPayload{User: user, Subject: subject, Content: content})
	if err != nil {
		return err
	}

	task := asynq.NewTask(tasks.TypeEmail, payload)
	info, err := client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(3*time.Minute), asynq.Deadline(time.Now().Add(20*time.Minute)))

	notification_logger.WithFields(logrus.Fields{
		"id": string(info.ID),

		"queue": info.Queue,
	}).Info("Added email task to queue")

	return err
}

func HandleTicketAllocationAddToQueue(db *gorm.DB, ticketId int) error {
	/*
		Handler that when a ticket allocation is created, it adds a job to the queue
		to send an email to the user that the ticket allocation was created.
	*/
	var ticket models.Ticket
	err := db.
		Preload("TicketRequest.User").
		Preload("TicketRequest.TicketRelease.Event.Organization").First(&ticket, ticketId).Error
	if err != nil {
		return err
	}

	user := ticket.TicketRequest.User
	ticketRelease := ticket.TicketRequest.TicketRelease
	event := ticketRelease.Event

	if user.Email == "" {
		return fmt.Errorf("user email is empty")
	}

	var payWithin int = 0
	if ticketRelease.PayWithin != nil {
		payWithin = int(*ticketRelease.PayWithin)
	}

	var data = types.EmailTicketAllocationCreated{
		FullName:          user.FullName(),
		EventName:         event.Name,
		TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
		OrganizationName:  event.Organization.Name,
		OrganizationEmail: event.Organization.Email,
		PayWithin:         fmt.Sprintf("%d", payWithin),
	}

	tmpl, err := template.ParseFiles("templates/emails/ticket_allocation_created.html")
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}

	// The HTML content of the email is now in buf
	htmlContent := buf.String()

	htmlContent = strings.ReplaceAll(htmlContent, "\x00", "")

	// Create the data to be sent
	// TODO add event
	AddEmailJobToQueue(db, &user, "Your ticket to", htmlContent, nil)

	return nil
}

func HandleEmailJob(db *gorm.DB) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p tasks.EmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		var notification = models.Notification{
			UserUGKthID: p.User.UGKthID,
			Type:        models.EmailNotification,
			Subject:     p.Subject,
			Content:     p.Content,
			EventID:     p.EventID,
		}

		if err := notification.Validate(); err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"error":        err,
			}).Error("Error validating notification")

			return err
		}

		err := SendEmail(p.User, p.Subject, p.Content)
		if err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"email":        p.User.Email,
				"error":        err,
			}).Error("Error sending email")

			notification.Status = models.NotificationStatusFailed

			CreateNotificationInDb(db, &notification)

			return err
		}
		// It was a success, create notification in db
		notification.Status = models.NotificationStatusSent

		CreateNotificationInDb(db, &notification)

		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"email":        p.User.Email,
		}).Info("Email sent successfully")

		return nil
	}
}
