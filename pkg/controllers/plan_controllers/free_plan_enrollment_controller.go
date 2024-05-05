package plan_controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	plan_enrollment_service "github.com/DowLucas/gin-ticket-release/pkg/services/plan_enrollment"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	response_utils "github.com/DowLucas/gin-ticket-release/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FreePlanEnrollmentController struct {
	DB      *gorm.DB
	service *plan_enrollment_service.FreePlanEnrollmentService
}

// NewFreePlanEnrollmentController creates a new controller with the given database client
func NewFreePlanEnrollmentController(db *gorm.DB) *FreePlanEnrollmentController {
	return &FreePlanEnrollmentController{DB: db, service: plan_enrollment_service.NewFreePlanEnrollmentService(db)}
}

// Define the request data

func (fpec *FreePlanEnrollmentController) EnrollFreePlan(c *gin.Context) {
	// Enroll the user in the free plan

	user := c.MustGet("user").(models.User)

	var body types.FreeEnrollmentPlanBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response_utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Hand over to service
	err := fpec.service.Enroll(&user, body)
	if err != nil {
		response_utils.RespondWithError(c, err.StatusCode, err.Message)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User enrolled in free plan"})
}
