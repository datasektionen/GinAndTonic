package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketReleaseMethodsController struct {
	DB *gorm.DB
}

func NewTicketReleaseMethodsController(db *gorm.DB) *TicketReleaseMethodsController {
	return &TicketReleaseMethodsController{DB: db}
}

func (trmc *TicketReleaseMethodsController) CreateTicketReleaseMethod(c *gin.Context) {
	var trm models.TicketReleaseMethod

	if err := c.ShouldBindJSON(&trm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := trmc.DB.Create(&trm).Error; err != nil {
		utils.HandleDBError(c, err, "creating the ticket release method")
		return
	}

	c.JSON(http.StatusCreated, trm)
}

func (trmc *TicketReleaseMethodsController) ListTicketReleaseMethods(c *gin.Context) {
	var ticketReleaseMethods []models.TicketReleaseMethod
	if err := trmc.DB.Find(&ticketReleaseMethods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error listing the ticket release methods"})
		return
	}

	c.JSON(http.StatusOK, ticketReleaseMethods)
}
