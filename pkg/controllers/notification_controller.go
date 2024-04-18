package controllers

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationController struct {
	DB *gorm.DB
}

func NewNotificationController(db *gorm.DB) *NotificationController {
	return &NotificationController{DB: db}
}

type EmailRequest struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

func (nc *NotificationController) SendTestEmail(c *gin.Context) {
	/*
		Handler that when a ticket allocation is created, it adds a job to the queue
		to send an email to the user that the ticket allocation was created.
	*/
	var user models.User
	if err := nc.DB.First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// updatedTime := time.Now()
	// payWithin := 1

	// PayBefo
	// PayBefore := utils.ConvertPayWithinToString(payWithin, updatedTime)

	// var data = types.EmailTicketAllocationCreated{
	// 	FullName:          "Tsst",
	// 	EventName:         "Blums",
	// 	TicketURL:         os.Getenv("FRONTEND_BASE_URL") + "/profile/tickets",
	// 	TeamName:  "DKM",
	// 	TeamEmail: "test@datasektionen.se",
	// 	PayBefore:         PayBefore,
	// }

	var tickets []types.EmailTicket
	tickets = append(tickets, types.EmailTicket{
		Name:  "Test",
		Price: "100.00",
	})

	str, _ := utils.GenerateEmailTable(tickets)

	var data = types.EmailTicketNotPaidInTime{
		FullName:    "Tsst",
		EventName:   "Blums",
		TicketsHTML: template.HTML(str),
		TeamEmail:   "test@gmail.com",
	}

	tmpl, err := template.ParseFiles("templates/emails/ticket_not_paid_in_time.html")
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
	jobs.AddEmailJobToQueue(db, &user, "Your ticket to", htmlContent, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}
