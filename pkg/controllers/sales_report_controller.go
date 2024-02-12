package controllers

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
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
