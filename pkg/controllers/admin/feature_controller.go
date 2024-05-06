package admin_controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FeatureController struct {
	DB *gorm.DB
}

func NewFeatureController(db *gorm.DB) *FeatureController {
	return &FeatureController{DB: db}
}

func (ctrl *FeatureController) GetAllFeatures(c *gin.Context) {
	queryParams, err := utils.GetQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	var features []models.Feature
	if err := ctrl.DB.Order(sort + " " + order).Offset(queryParams.Range[0]).Limit(queryParams.Range[1] - queryParams.Range[0] + 1).Find(&features).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var count int64
	ctrl.DB.Model(&models.Feature{}).Count(&count)

	c.Header("X-Total-Count", fmt.Sprintf("%d", count))
	c.Header("Content-Range", fmt.Sprintf("features %d-%d/%d", queryParams.Range[0], queryParams.Range[1], count))

	c.JSON(http.StatusOK, features)
}
func (ctrl *FeatureController) GetFeature(c *gin.Context) {
	id := c.Param("id")
	var feature models.Feature
	if err := ctrl.DB.Preload("FeatureLimits").Preload("PackageTiers").First(&feature, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	tierIDs := make([]uint, len(feature.PackageTiers))
	for i, tier := range feature.PackageTiers {
		tierIDs[i] = tier.ID
	}

	feature.PackageTiersIDs = tierIDs

	c.JSON(http.StatusOK, feature)
}

func (fc *FeatureController) CreateFeature(c *gin.Context) {
	var feature models.Feature
	if err := c.ShouldBindJSON(&feature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Start a transaction
	tx := fc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&feature).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(feature.PackageTiersIDs) > 0 {
		var tiers []models.PackageTier
		if err := tx.Find(&tiers, feature.PackageTiersIDs).Error; err == nil {
			tx.Model(&feature).Association("PackageTiers").Replace(tiers)
		}
	}

	tx.Commit()
	c.JSON(http.StatusCreated, feature)
}

func (ctrl *FeatureController) UpdateFeature(c *gin.Context) {
	id := c.Param("id")
	tx := ctrl.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var feature models.Feature
	if err := tx.First(&feature, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err := c.ShouldBindJSON(&feature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(len(feature.FeatureLimits))

	// Delete Unscoped all current feature limits
	if err := tx.Unscoped().Where("feature_id = ?", feature.ID).Delete(&models.FeatureLimit{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	packageTiersIds := make([]uint, len(feature.PackageTiers))

	for _, limit := range feature.FeatureLimits {
		packageTierId := limit.PackageTierID
		packageTiersIds = append(packageTiersIds, packageTierId)
	}

	var tiers []models.PackageTier
	if err := tx.Find(&tiers, packageTiersIds).Error; err == nil {
		tx.Model(&feature).Association("PackageTiers").Replace(tiers)
	} else {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Save(&feature).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	if tx.Error != nil {
		fmt.Println(tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error})
		return
	}

	c.JSON(http.StatusOK, feature)
}

func (ctrl *FeatureController) DeleteFeature(c *gin.Context) {
	id := c.Param("id")

	val, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = models.DeleteFeature(ctrl.DB, uint(val))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
