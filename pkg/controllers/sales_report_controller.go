package controllers

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services/aws_service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SalesReportController struct {
	DB *gorm.DB
}

func NewSalesReportController(db *gorm.DB) *SalesReportController {
	return &SalesReportController{DB: db}
}

func (src *SalesReportController) GenerateSalesReport(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = jobs.AddSalesReportJobToQueue(eventID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sales report generation job added to queue"})
}

func (src *SalesReportController) ListSalesReport(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var salesReports []models.EventSalesReport
	if err := src.DB.Where("event_id = ?", eventID).Find(&salesReports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s3Client, err := aws_service.NewS3Client()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range salesReports {
		url, err := aws_service.GetFileURL(s3Client, salesReports[i].FileName)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		salesReports[i].URL = url
	}

	c.JSON(http.StatusOK, salesReports)
}
