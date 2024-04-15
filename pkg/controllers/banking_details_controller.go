package controllers

import (
	"net/http"
	"strconv"

	services "github.com/DowLucas/gin-ticket-release/pkg/services/banking"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
)

type BankingController struct {
	service *services.BankingService
}

func NewBankingController(service *services.BankingService) *BankingController {
	return &BankingController{
		service: service,
	}
}

func (bc *BankingController) SubmitBankingDetails(c *gin.Context) {
	organizationIDstring := c.Param("organizationID")
	if organizationIDstring == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id is required"})
		return
	}

	organizationID, err := strconv.Atoi(organizationIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id must be an integer"})
		return
	}

	var details types.BankingDetailsRequest

	if err := c.ShouldBindJSON(&details); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := bc.service.SubmitBankingDetails(&details, uint(organizationID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Banking details submitted successfully"})
}

func (bc *BankingController) GetBankingDetails(c *gin.Context) {
	organizationIDstring := c.Param("organizationID")
	if organizationIDstring == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id is required"})
		return
	}

	organizationID, err := strconv.Atoi(organizationIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id must be an integer"})
		return
	}

	details, rerr := bc.service.GetBankingDetails(uint(organizationID))
	if rerr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": rerr.Message})
		return
	}

	c.JSON(http.StatusOK, details)
}

func (bc *BankingController) DeleteBankingDetails(c *gin.Context) {
	organizationIDstring := c.Param("organizationID")
	if organizationIDstring == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id is required"})
		return
	}

	organizationID, err := strconv.Atoi(organizationIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id must be an integer"})
		return
	}

	rerr := bc.service.DeleteBankingDetails(uint(organizationID))
	if rerr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": rerr.Message})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
