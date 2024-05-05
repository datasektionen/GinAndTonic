package routes

import (
	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers/plan_controllers"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PlanEnrollmentRoutes(r *gin.Engine, db *gorm.DB) *gin.Engine {

	freePlanEnrollmentController := plan_controllers.NewFreePlanEnrollmentController(db)

	planEnrollmentGroup := r.Group("/plan-enrollments")
	planEnrollmentGroup.Use(authentication.ValidateTokenMiddleware(true))
	planEnrollmentGroup.Use(middleware.UserLoader(db))

	planEnrollmentGroup.POST("/free", freePlanEnrollmentController.EnrollFreePlan)

	return r
}
