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

type PlanEnrollmentAdminController struct {
	DB      *gorm.DB
	service *admin_services.PlanEnrollmentAdminService
}

// NewPlanEnrollmentAdminController creates a new controller with the given database client
func NewPlanEnrollmentAdminController(db *gorm.DB) *PlanEnrollmentAdminController {
	return &PlanEnrollmentAdminController{DB: db, service: admin_services.NewPlanEnrollmentAdminService(db)}
}

// GetAllPackages fetches all pricing packages
// GetAllPackages fetches all pricing packages with query params handling
func (pc *PlanEnrollmentAdminController) GetAllEnrollments(c *gin.Context) {
	queryParams, err := utils.GetQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters", "details": err.Error()})
		return
	}

	var enrollments []models.PlanEnrollment
	query := pc.DB.Model(&models.PlanEnrollment{})

	sortParam := c.DefaultQuery("sort", "id")
	sortArray := strings.Split(strings.Trim(sortParam, "[]\""), "\",\"")
	sort := sortArray[0]
	order := "asc" // default order
	if len(sortArray) > 1 {
		order = sortArray[1]
	}

	// Execute query
	if result := query.Order(fmt.Sprintf("%s %s", sort, order)).Offset(queryParams.Range[0]).Limit(queryParams.Range[1] - queryParams.Range[0] + 1).Find(&enrollments); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Assuming total count for headers (pagination)
	var totalCount int64
	pc.DB.Model(&models.PlanEnrollment{}).Count(&totalCount)
	c.Header("X-Total-Count", fmt.Sprintf("%d", totalCount))
	c.Header("Content-Range", fmt.Sprintf("enrollments %d-%d/%d", queryParams.Range[0], queryParams.Range[1], totalCount))

	c.JSON(http.StatusOK, enrollments)
}

// GetPackage fetches a single pricing package by ID
func (pc *PlanEnrollmentAdminController) GetEnrollment(c *gin.Context) {
	id := c.Param("id")
	var planEnrollment models.PlanEnrollment
	if result := pc.DB.Preload("Features").First(&planEnrollment, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, planEnrollment)
}

type CreatePackageRequest struct {
	PlanEnrollment models.PlanEnrollment `json:"pricing_package"`
}

func (pc *PlanEnrollmentAdminController) CreateEnrollment(c *gin.Context) {
	var planEnrollment models.PlanEnrollment
	// Print the body

	if err := c.ShouldBindJSON(&planEnrollment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if result := pc.DB.Create(&planEnrollment); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, planEnrollment)
}

// UpdatePackage updates an existing pricing package
func (pc *PlanEnrollmentAdminController) UpdateEnrollment(c *gin.Context) {
	id := c.Param("id")
	var existingPlanEnrollment models.PlanEnrollment
	if result := pc.DB.First(&existingPlanEnrollment, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Package not found"})
		return
	}

	var body models.PlanEnrollment
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.PackageTierID != existingPlanEnrollment.PackageTierID {
		var newTier models.PackageTier
		if result := pc.DB.First(&newTier, body.PackageTierID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
			return
		}

		if err := existingPlanEnrollment.ClearFeatures(pc.DB); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error clearing default features"})
			return
		}

		// Get the default features that come with the new tier
		defaultFeatures, err := newTier.GetDefaultFeatures(pc.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting default features"})
			return
		}

		// Update the features
		existingPlanEnrollment.Features = defaultFeatures
	}

	existingPlanEnrollment.ReferenceName = body.ReferenceName
	existingPlanEnrollment.PackageTierID = body.PackageTierID
	existingPlanEnrollment.MonthlyPrice = body.MonthlyPrice
	existingPlanEnrollment.YearlyPrice = body.YearlyPrice
	existingPlanEnrollment.OneTimePrice = body.OneTimePrice
	existingPlanEnrollment.Plan = body.Plan

	pc.DB.Save(&existingPlanEnrollment)

	c.JSON(http.StatusOK, existingPlanEnrollment)
}

// DeletePackage deletes a pricing package
func (pc *PlanEnrollmentAdminController) DeleteEnrollment(c *gin.Context) {
	id := c.Param("id")
	if result := pc.DB.Delete(&models.PlanEnrollment{}, id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
