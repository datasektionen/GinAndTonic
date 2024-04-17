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
	"github.com/DowLucas/gin-ticket-release/utils"
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

// CreateNotificationInDb creates a notification in the database
func CreateNotificationInDb(tx *gorm.DB, notification *models.Notification) error {
	// Create notification
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"error":        err,
		}).Error("Error creating notification in database")

		return err
	}

	return nil
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

func AddEmailJobToQueue(db *gorm.DB, user *models.User, subject, content string, eventId *uint) error {
	client := connectAsynqClient()
	defer client.Close()

	payload, err := json.Marshal(tasks.EmailPayload{User: user, Subject: subject, Content: content, EventID: eventId})
	if err != nil {
		return err
	}

	// Calculate the schedule time
	deadline := time.Now().Add(10 * time.Minute)

	task := asynq.NewTask(tasks.TypeEmail, payload)
	info, err := client.Enqueue(task,
		asynq.Queue("email"),
		asynq.MaxRetry(3),
		asynq.Deadline(deadline),
		asynq.Timeout(3*time.Minute))

	if err != nil {
		return err
	}

	notification_logger.WithFields(logrus.Fields{
		"id":    string(info.ID),
		"queue": info.Queue,
	}).Info("Added email task to queue")

	return err
}

func AddReminderEmailJobToQueueAt(db *gorm.DB, user *models.User,
	subject, content string, reminderId uint, scheduleTime time.Time) error {
	client := connectAsynqClient()
	defer client.Close()

	payload, err := json.Marshal(tasks.EmailReminderPayload{User: user, Subject: subject, Content: content, ReminderID: reminderId})
	if err != nil {
		return err
	}

	// Calculate the delay
	delay := scheduleTime.Sub(time.Now())

	task := asynq.NewTask(tasks.TypeReminderEmail, payload)
	info, err := client.Enqueue(task, asynq.Queue("email"), asynq.MaxRetry(3), asynq.ProcessIn(delay))

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
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		var p tasks.EmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		content, err := utils.CompressHTML(p.Content)
		if err != nil {
			notification_logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("Error compressing HTML")
			content = p.Content
		}

		var notification = models.Notification{
			UserUGKthID: p.User.UGKthID,
			Type:        models.EmailNotification,
			Subject:     &p.Subject,
			EventID:     p.EventID,
			Content:     &content,
			Status:      models.PendingNotification,
		}

		if err := notification.Validate(); err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"error":        err,
			}).Error("Error validating notification")

			return err
		}

		err = SendEmail(p.User, p.Subject, p.Content, tx)
		if err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"email":        p.User.Email,
				"error":        err,
			}).Error("Error sending email")

			notification.Status = models.FailedNotification

			CreateNotificationInDb(tx, &notification)

			return err
		}
		// It was a success, create notification in db
		notification.Status = models.SentNotification

		CreateNotificationInDb(tx, &notification)

		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"email":        p.User.Email,
		}).Info("Email sent successfully")

		if err := tx.Commit().Error; err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"error":        err,
			}).Error("Error committing transaction")
			tx.Rollback()

			return err
		}

		return nil
	}
}

func HandleReminderJob(db *gorm.DB) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		var p tasks.EmailReminderPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		var ticketReleaseReminder *models.TicketReleaseReminder
		if err := db.Preload("TicketRelease").Where("id = ?", p.ReminderID).First(&ticketReleaseReminder).Error; err != nil {
			notification_logger.WithFields(logrus.Fields{
				"reminder_id": p.ReminderID,
			}).Error("Reminder has been deleted")

			return nil
		}

		// Get the open of the ticket release
		opensAt := time.Unix(ticketReleaseReminder.TicketRelease.Open, 0)

		// If the ticket release open is more than 10 minutes from now, we should not send the reminder
		// Instead we should schedule a new reminder

		if opensAt.Sub(time.Now()) > 10*time.Minute {
			notification_logger.WithFields(logrus.Fields{
				"reminder_id": p.ReminderID,
				"opens_at":    opensAt,
			}).Info("Ticket release opens more than 10 minutes from now, scheduling a new reminder")

			newReminder := models.TicketReleaseReminder{
				TicketReleaseID: ticketReleaseReminder.TicketReleaseID,
				UserUGKthID:     ticketReleaseReminder.UserUGKthID,
				ReminderTime:    opensAt.Add(-10 * time.Minute),
				IsSent:          false,
			}

			if err := tx.Create(&newReminder).Error; err != nil {
				notification_logger.WithFields(logrus.Fields{
					"reminder_id": p.ReminderID,
					"error":       err,
				}).Error("Error creating new reminder")

				return err
			}

			// Schedule a new reminder
			err := AddReminderEmailJobToQueueAt(tx, p.User, p.Subject, p.Content, newReminder.ID, newReminder.ReminderTime)
			if err != nil {
				notification_logger.WithFields(logrus.Fields{
					"reminder_id": p.ReminderID,
					"error":       err,
				}).Error("Error scheduling new reminder")

				return err
			}

			return nil
		}

		var notification = models.Notification{
			UserUGKthID: p.User.UGKthID,
			Type:        models.EmailNotification,
			Subject:     &p.Subject,
			Content:     &p.Content,
		}

		CreateNotificationInDb(tx, &notification)

		if err := notification.Validate(); err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"error":        err,
			}).Error("Error validating notification")

			if err := handleFailedNotification(&notification, err, tx); err != nil {
				notification_logger.WithFields(logrus.Fields{
					"notification": notification,
					"error":        err,
				}).Error("Error handling failed notification")
			}

			tx.Rollback()

			return err
		}

		err := SendEmail(p.User, p.Subject, p.Content, tx)
		if err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"email":        p.User.Email,
				"error":        err,
			}).Error("Error sending email")

			if err := handleFailedNotification(&notification, err, tx); err != nil {
				notification_logger.WithFields(logrus.Fields{
					"notification": notification,
					"error":        err,
				}).Error("Error handling failed notification")
			}

			tx.Rollback()

			return err
		}

		ticketReleaseReminder.IsSent = true
		if err := db.Save(&ticketReleaseReminder).Error; err != nil {
			notification_logger.WithFields(logrus.Fields{
				"reminder_id": p.ReminderID,
				"error":       err,
			}).Error("Error saving ticket release reminder")

			return err
		}

		if err := setSentNotification(&notification, tx); err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"error":        err,
			}).Error("Error setting notification as sent")

			if err := handleFailedNotification(&notification, err, tx); err != nil {
				notification_logger.WithFields(logrus.Fields{
					"notification": notification,
					"error":        err,
				}).Error("Error handling failed notification")
			}

			tx.Rollback()
			return err
		}

		if err := tx.Commit().Error; err != nil {
			notification_logger.WithFields(logrus.Fields{
				"notification": notification,
				"error":        err,
			}).Error("Error committing transaction")
			tx.Rollback()

			return err
		}

		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"email":        p.User.Email,
		}).Info("Email sent successfully")

		return nil
	}
}
