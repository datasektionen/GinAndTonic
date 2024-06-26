package jobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

// MailData is the data that is sent to the spam API
type MailData struct {
	Key     string `json:"key"`
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Content string `json:"content"`
	ReplyTo string `json:"replyTo"`
}

// From is the email address that the emails will be sent from
const From = "tessera-no-reply@datasektionen.se"

// SpamURL is the URL to the spam API
const SpamURL = "https://spam.datasektionen.se/api/sendmail"

// SendContactEmail sends an email to the contact email
func SendContactEmail(name, email_to, from, subject, content string) error {
	// Create the data to be sent
	var to string
	if os.Getenv("ENV") == "dev" {
		to = os.Getenv("SPAM_TEST_EMAIL")
	} else {
		to = email_to
	}

	data := MailData{
		Key:     os.Getenv("SPAM_API_KEY"),
		To:      to,
		From:    "tessera-contact@datasektionen.se",
		Subject: subject,
		Content: content,
		ReplyTo: from,
	}

	// Marshal the data into a JSON payload
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payloadBuffer := bytes.NewBuffer(payloadBytes)

	// Create a new request
	req, err := http.NewRequest("POST", SpamURL, payloadBuffer)
	if err != nil {
		return err
	}

	// Set the appropriate headers (Content-Type)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// SendEmail sends an email to the user
func SendEmail(user *models.User, subject, content string, db *gorm.DB) error {
	// Create the data to be sent
	var to string
	if os.Getenv("ENV") == "dev" {
		to = os.Getenv("SPAM_TEST_EMAIL")
	} else {
		to = user.GetUserEmail(db)
	}

	data := MailData{
		Key:     os.Getenv("SPAM_API_KEY"),
		To:      to,
		From:    From,
		Subject: subject,
		Content: content,
	}

	// Marshal the data into a JSON payload
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payloadBuffer := bytes.NewBuffer(payloadBytes)

	// Create a new request
	req, err := http.NewRequest("POST", SpamURL, payloadBuffer)
	if err != nil {
		return err
	}

	// Set the appropriate headers (Content-Type)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
