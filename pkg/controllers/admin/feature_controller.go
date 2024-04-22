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
	sort := sortArray[0]
	order := sortArray[1]

	var features []models.Feature
	if err := ctrl.DB.Order(sort + " " + order).Offset((queryParams.Page - 1) * queryParams.PerPage).Limit(queryParams.PerPage).Find(&features).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var count int64
	ctrl.DB.Model(&models.Feature{}).Count(&count)

	c.Header("X-Total-Count", fmt.Sprintf("%d", count))
	c.Header("Content-Range", fmt.Sprintf("features %d-%d/%d", (queryParams.Page-1)*queryParams.PerPage, (queryParams.Page-1)*queryParams.PerPage+len(features)-1, count))

	c.JSON(http.StatusOK, features)
}

func (ctrl *FeatureController) GetFeature(c *gin.Context) {
	id := c.Param("id")
	var feature models.Feature
	if err := ctrl.DB.Preload("FeatureLimit").First(&feature, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	c.JSON(http.StatusOK, feature)
}

func (ctrl *FeatureController) CreateFeature(c *gin.Context) {
	var feature models.Feature
	if err := c.ShouldBindJSON(&feature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctrl.DB.Create(&feature).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, feature)
}

func (ctrl *FeatureController) UpdateFeature(c *gin.Context) {
	id := c.Param("id")
	var feature models.Feature
	if err := ctrl.DB.First(&feature, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err := c.ShouldBindJSON(&feature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctrl.DB.Save(&feature)
	c.JSON(http.StatusOK, feature)
}

func (ctrl *FeatureController) DeleteFeature(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.DB.Delete(&models.Feature{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
