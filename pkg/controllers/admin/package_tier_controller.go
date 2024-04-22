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

	validSortColumns := []string{"id", "name", "tier", "description", "standard_monthly_price", "standard_yearly_price"}
	sortParam := c.DefaultQuery("sort", "id")
	sortArray := strings.Split(strings.Trim(sortParam, "[]\""), "\",\"")
	var sort, order string
	if len(sortArray) == 2 {
		sort = sortArray[0]
		order = sortArray[1]
	} else {
		sort = "id"
		order = "asc"
	}

	isValidColumn := false
	for _, column := range validSortColumns {
		if sort == column {
			isValidColumn = true
			break
		}
	}
	if !isValidColumn {
		sort = "id"
	}

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
	if result := ptc.DB.Preload("DefaultFeatures").First(&tier, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		return
	}

	featureIDs := make([]uint, len(tier.DefaultFeatures))
	for i, feature := range tier.DefaultFeatures {
		featureIDs[i] = feature.ID
	}

	tier.DefaultFeatureIDs = featureIDs

	c.JSON(http.StatusOK, tier)
}

// CreateTier creates a new package tier
func (ptc *PackageTierController) CreateTier(c *gin.Context) {
	var tier models.PackageTier
	if err := c.ShouldBindJSON(&tier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the tier without associated features
	if result := ptc.DB.Omit("DefaultFeatures.*").Create(&tier); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Manually managing the many-to-many relation
	if len(tier.DefaultFeatures) > 0 {
		if err := ptc.DB.Model(&tier).Association("DefaultFeatures").Replace(tier.DefaultFeatures); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, tier)
}

// UpdateTier updates an existing package tier
func (ptc *PackageTierController) UpdateTier(c *gin.Context) {
	id := c.Param("id")
	var tier models.PackageTier
	if result := ptc.DB.Preload("DefaultFeatures").First(&tier, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		return
	}

	if err := c.ShouldBindJSON(&tier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the tier without associated features to avoid complications with JSON binding
	if result := ptc.DB.Save(&tier); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if len(tier.DefaultFeatureIDs) > 0 {
		var features []models.Feature
		if err := ptc.DB.Find(&features, tier.DefaultFeatureIDs).Error; err == nil {
			ptc.DB.Model(&tier).Association("DefaultFeatures").Replace(features)
		}
	}

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
