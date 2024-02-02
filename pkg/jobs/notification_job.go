package jobs

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs/tasks"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var notification_logger = logrus.New()

func init() {
	// Load or create log file
	// Create logs directory if it doesn't exist
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	if _, err := os.Stat("logs/allocate_reserve_tickets_job.log"); os.IsNotExist(err) {
		os.Create("logs/allocate_reserve_tickets_job.log")
	}

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

func connectAsynqClient() *asynq.Client {
	// Parse the REDIS_URL
	redisURL, err := url.Parse(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("Failed to parse REDIS_URL: %v", err)
	}

	// Extract host and password. Port is part of the host in the URL.
	redisHost := redisURL.Host
	redisPassword, _ := redisURL.User.Password()

	// Create a new Asynq client instance with RedisClientOpt.
	var client *asynq.Client

	if os.Getenv("ENV") == "dev" {
		client = asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")})
	} else {
		client = asynq.NewClient(asynq.RedisClientOpt{Addr: redisHost, Password: redisPassword})
	}

	return client
}

func AddEmailJobToQueue(db *gorm.DB, user *models.User, subject, content string, evnetId *uint) error {
	client := connectAsynqClient()
	defer client.Close()

	payload, err := json.Marshal(tasks.EmailPayload{User: user, Subject: subject, Content: content, EventID: evnetId})
	if err != nil {
		return err
	}

	task := asynq.NewTask(tasks.TypeEmail, payload)
	info, err := client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(3*time.Minute), asynq.Deadline(time.Now().Add(20*time.Minute)))

	if err != nil {
		return err
	}

	notification_logger.WithFields(logrus.Fields{
		"id":    string(info.ID),
		"queue": info.Queue,
	}).Info("Added email task to queue")

	return err
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
