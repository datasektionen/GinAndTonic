package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs/tasks"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func AddSendOutEmailJobToQueue(db *gorm.DB, user *models.User, sendOut *models.SendOut, message string) error {
	client := connectAsynqClient()
	defer client.Close()

	payload, err := json.Marshal(tasks.SendOutEmailPayload{User: user, SendOut: sendOut, Content: message})
	if err != nil {
		return err
	}

	// Calculate the schedule time
	deadline := time.Now().Add(10 * time.Minute)

	task := asynq.NewTask(tasks.TypeSendOutEmail, payload)
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

func HandleSendOutEmailJob(db *gorm.DB) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		var p tasks.SendOutEmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		user := p.User
		sendOut := p.SendOut

		notification, err := createSendOutNotification(user, sendOut, tx)
		if err != nil {
			return err
		}

		err = SendEmail(user, sendOut.Subject, p.Content, tx)
		if err != nil {
			err := handleFailedNotification(notification, err, tx)
			if err != nil {
				return err
			}
		}

		err = setSentNotification(notification, tx)
		if err != nil {
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
			"email":        user.Email,
		}).Info("Email sent successfully")

		return nil
	}
}

func createSendOutNotification(user *models.User, sendOut *models.SendOut, tx *gorm.DB) (*models.Notification, error) {
	notification := models.Notification{
		UserUGKthID: user.UGKthID,
		Type:        models.EmailNotification,
		EventID:     sendOut.EventID,
		Status:      models.PendingNotification,
		SendOutID:   &sendOut.ID,
	}

	if err := notification.Validate(); err != nil {
		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"error":        err,
		}).Error("Error validating notification")

		return nil, err
	}

	err := CreateNotificationInDb(tx, &notification)
	if err != nil {
		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"error":        err,
		}).Error("Error creating notification")
		tx.Rollback()
		return nil, err
	}

	return &notification, nil
}

func handleFailedNotification(notification *models.Notification, err error, tx *gorm.DB) error {
	notification_logger.WithFields(logrus.Fields{
		"notification": notification,
		"error":        err,
	}).Error("Error sending email")

	err = notification.SetFailed(tx, err)
	if err != nil {
		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"error":        err,
		}).Error("Error setting notification to failed")
		tx.Rollback()

		return err
	}

	return nil
}

func setSentNotification(notification *models.Notification, tx *gorm.DB) error {
	err := notification.SetSent(tx)
	if err != nil {
		notification_logger.WithFields(logrus.Fields{
			"notification": notification,
			"error":        err,
		}).Error("Error setting notification to sent")
		tx.Rollback()
		return err
	}

	return nil
}
