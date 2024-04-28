package admin_controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	admin_services "github.com/DowLucas/gin-ticket-release/pkg/services/admin"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PricingPackageAdminController struct {
	DB      *gorm.DB
	service *admin_services.PricingPackageAdminService
}

// NewPricingPackageAdminController creates a new controller with the given database client
func NewPricingPackageAdminController(db *gorm.DB) *PricingPackageAdminController {
	return &PricingPackageAdminController{DB: db, service: admin_services.NewPricingPackageAdminService(db)}
}

// GetAllPackages fetches all pricing packages
// GetAllPackages fetches all pricing packages with query params handling
func (pc *PricingPackageAdminController) GetAllPackages(c *gin.Context) {
	queryParams, err := utils.GetQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters", "details": err.Error()})
		return
	}

	var packages []models.PricingPackage
	query := pc.DB.Model(&models.PricingPackage{})

	sortParam := c.DefaultQuery("sort", "id")
	sortArray := strings.Split(strings.Trim(sortParam, "[]\""), "\",\"")
	sort := sortArray[0]
	order := sortArray[1]

	// Execute query
	if result := query.Order(fmt.Sprintf("%s %s", sort, order)).Offset((queryParams.Page - 1) * queryParams.PerPage).Limit(queryParams.PerPage).Find(&packages); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	var count int64
	pc.DB.Model(&models.PricingPackage{}).Count(&count)

	// Assuming total count for headers (pagination)
	var totalCount int64
	pc.DB.Model(&models.PricingPackage{}).Count(&totalCount)
	c.Header("X-Total-Count", fmt.Sprintf("%d", totalCount))
	c.Header("Content-Range", fmt.Sprintf("packages %d-%d/%d", (queryParams.Page-1)*queryParams.PerPage, (queryParams.Page-1)*queryParams.PerPage+len(packages)-1, count))

	c.JSON(http.StatusOK, packages)
}

// GetPackage fetches a single pricing package by ID
func (pc *PricingPackageAdminController) GetPackage(c *gin.Context) {
	id := c.Param("id")
	var pricingPackage models.PricingPackage
	if result := pc.DB.First(&pricingPackage, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, pricingPackage)
}
func (pc *PricingPackageAdminController) CreatePackage(c *gin.Context) {
	var pricingPackage models.PricingPackage
	// Print the body

	if err := c.ShouldBindJSON(&pricingPackage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if result := pc.DB.Create(&pricingPackage); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, pricingPackage)
}

// UpdatePackage updates an existing pricing package
func (pc *PricingPackageAdminController) UpdatePackage(c *gin.Context) {
	id := c.Param("id")
	var pricingPackage models.PricingPackage
	if result := pc.DB.First(&pricingPackage, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Package not found"})
		return
	}

	if err := c.ShouldBindJSON(&pricingPackage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pc.DB.Save(&pricingPackage)

	c.JSON(http.StatusOK, pricingPackage)
}

// DeletePackage deletes a pricing package
func (pc *PricingPackageAdminController) DeletePackage(c *gin.Context) {
	id := c.Param("id")
	if result := pc.DB.Delete(&models.PricingPackage{}, id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
