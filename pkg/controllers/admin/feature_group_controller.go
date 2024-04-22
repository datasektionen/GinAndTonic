package admin_controllers

import (
	"fmt"
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FeatureGroupController struct {
	DB *gorm.DB
}

func NewFeatureGroupController(db *gorm.DB) *FeatureGroupController {
	return &FeatureGroupController{DB: db}
}

func (fgc *FeatureGroupController) GetAllFeatureGroups(c *gin.Context) {
	var featureGroups []models.FeatureGroup
	if err := fgc.DB.Find(&featureGroups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var count int64
	fgc.DB.Model(&models.FeatureGroup{}).Count(&count)

	c.Header("X-Total-Count", fmt.Sprintf("%d", count))
	c.Header("Content-Range", fmt.Sprintf("featureGroups 0-%d/%d", len(featureGroups)-1, count))

	c.JSON(http.StatusOK, featureGroups)
}

func (fgc *FeatureGroupController) GetFeatureGroup(c *gin.Context) {
	id := c.Param("id")
	var featureGroup models.FeatureGroup
	if err := fgc.DB.First(&featureGroup, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature group not found"})
		return
	}
	c.JSON(http.StatusOK, featureGroup)
}

func (fgc *FeatureGroupController) CreateFeatureGroup(c *gin.Context) {
	var featureGroup models.FeatureGroup
	if err := c.ShouldBindJSON(&featureGroup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := fgc.DB.Create(&featureGroup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, featureGroup)
}

func (fgc *FeatureGroupController) UpdateFeatureGroup(c *gin.Context) {
	id := c.Param("id")
	var featureGroup models.FeatureGroup
	if err := fgc.DB.First(&featureGroup, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature group not found"})
		return
	}
	if err := c.ShouldBindJSON(&featureGroup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fgc.DB.Save(&featureGroup)
	c.JSON(http.StatusOK, featureGroup)
}

func (fgc *FeatureGroupController) DeleteFeatureGroup(c *gin.Context) {
	id := c.Param("id")
	if err := fgc.DB.Delete(&models.FeatureGroup{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
