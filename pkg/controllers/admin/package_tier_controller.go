package admin_controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PackageTierController struct {
	DB *gorm.DB
}

func NewPackageTierController(db *gorm.DB) *PackageTierController {
	return &PackageTierController{DB: db}
}

// GetAllTiers fetches all package tiers
// GetAllTiers fetches all package tiers
func (ptc *PackageTierController) GetAllTiers(c *gin.Context) {
	queryParams, err := utils.GetQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tiers []models.PackageTier

	sortParam := c.DefaultQuery("sort", "id")
	sortArray := strings.Split(strings.Trim(sortParam, "[]\""), "\",\"")
	sort := sortArray[0]
	order := sortArray[1]

	// Use the query parameters in the database query
	if result := ptc.DB.Order(fmt.Sprintf("%s %s", sort, order)).Offset((queryParams.Page - 1) * queryParams.PerPage).Limit(queryParams.PerPage).Find(&tiers); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	// Get the total count of tiers
	var count int64
	ptc.DB.Model(&models.PackageTier{}).Count(&count)

	// Set the Content-Range header
	c.Header("Content-Range", fmt.Sprintf("tiers %d-%d/%d", (queryParams.Page-1)*queryParams.PerPage, (queryParams.Page-1)*queryParams.PerPage+len(tiers)-1, count))

	c.JSON(http.StatusOK, tiers)
}

// GetTier fetches a single package tier by ID
func (ptc *PackageTierController) GetTier(c *gin.Context) {
	id := c.Param("id")
	var tier models.PackageTier
	if result := ptc.DB.First(&tier, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		return
	}
	c.JSON(http.StatusOK, tier)
}

// CreateTier creates a new package tier
func (ptc *PackageTierController) CreateTier(c *gin.Context) {
	var tier models.PackageTier
	if err := c.ShouldBindJSON(&tier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if result := ptc.DB.Create(&tier); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, tier)
}

// UpdateTier updates an existing package tier
func (ptc *PackageTierController) UpdateTier(c *gin.Context) {
	id := c.Param("id")
	var tier models.PackageTier
	if result := ptc.DB.First(&tier, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		return
	}

	if err := c.ShouldBindJSON(&tier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ptc.DB.Save(&tier)
	c.JSON(http.StatusOK, tier)
}

// DeleteTier deletes a package tier
func (ptc *PackageTierController) DeleteTier(c *gin.Context) {
	id := c.Param("id")
	if result := ptc.DB.Delete(&models.PackageTier{}, id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
