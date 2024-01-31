package jobs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

// create type for body
type MailData struct {
	Key     string `json:"key"`
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

const FROM = "tessera@datasektionen.se"

func SendEmail(user *models.User, subject, content string) error {
	// Create the data to be sent
	data := MailData{
		Key:     os.Getenv("SPAM_API_KEY"),
		To:      "lucdow7@gmail.com", // HARD CODED FOR TESTING TODO CHANGE
		From:    FROM,
		Subject: subject,
		Content: content,
	}

	// Marshal the data into a JSON payload
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payloadBuffer := bytes.NewBuffer(payloadBytes)

	// Define the URL
	url := "https://spam.datasektionen.se/api/sendmail"

	// Create a new request
	req, err := http.NewRequest("POST", url, payloadBuffer)
	if err != nil {
		return err
	}

	// Set the appropriate headers (Content-Type)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	// Read and print the response body
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}
