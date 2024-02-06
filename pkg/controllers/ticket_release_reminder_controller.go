package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketReleaseReminderController struct {
	DB *gorm.DB
}

// NewTicketReleaseReminderController creates a new controller with the given database client
func NewTicketReleaseReminderController(db *gorm.DB) *TicketReleaseReminderController {
	return &TicketReleaseReminderController{DB: db}
}

type CreateTicketReleaseReminderRequest struct {
	ReminderTime int `json:"reminder_time" binding:"required"`
}

func (erc *TicketReleaseReminderController) CreateTicketReleaseReminder(c *gin.Context) {
	ugkthid := c.MustGet("ugkthid").(string)

	ticketReleaseString := c.Param("ticketReleaseID")
	ticketReleaseID, err := strconv.Atoi(ticketReleaseString)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req CreateTicketReleaseReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var eventReminder models.TicketReleaseReminder = models.TicketReleaseReminder{
		TicketReleaseID: uint(ticketReleaseID),
		UserUGKthID:     ugkthid,
		ReminderTime:    time.Unix(int64(req.ReminderTime), 0),
	}

	if err := eventReminder.Validate(erc.DB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"validate error": err.Error()})
		return
	}

	if err := models.CreateTicketReleaseReminder(erc.DB, &eventReminder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"create": err.Error()})
		return
	}

	// No error, scheudle the reminder
	err = services.Notify_RemindUserOfTicketRelease(erc.DB, &eventReminder)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reminder error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, eventReminder)
}

func (erc *TicketReleaseReminderController) GetTicketReleaseReminder(c *gin.Context) {
	ugkthid := c.MustGet("ugkthid").(string)

	ticketReleaseString := c.Param("ticketReleaseID")
	ticketReleaseID, err := strconv.Atoi(ticketReleaseString)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var eventReminders models.TicketReleaseReminder
	if err := erc.DB.Where("ticket_release_id = ? AND user_ug_kth_id = ?", ticketReleaseID, ugkthid).First(&eventReminders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No reminder found"})
		return
	}

	c.JSON(http.StatusOK, eventReminders)
}

func (erc *TicketReleaseReminderController) DeleteTicketReleaseReminder(c *gin.Context) {
	ugkthid := c.MustGet("ugkthid").(string)

	ticketReleaseString := c.Param("ticketReleaseID")
	ticketReleaseID, err := strconv.Atoi(ticketReleaseString)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var eventReminders models.TicketReleaseReminder
	if err := erc.DB.Where("ticket_release_id = ? AND user_ug_kth_id = ?", ticketReleaseID, ugkthid).First(&eventReminders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No reminder found"})
		return
	}

	if err := erc.DB.Delete(&eventReminders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted"})
}
