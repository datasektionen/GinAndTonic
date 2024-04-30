package routes

import (
	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	admin_controllers "github.com/DowLucas/gin-ticket-release/pkg/controllers/admin"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminRoutes(r *gin.Engine, db *gorm.DB) *gin.Engine {
	PlanEnrollmentAdminController := admin_controllers.NewPlanEnrollmentAdminController(db)
	packageTierController := admin_controllers.NewPackageTierController(db)
	featureController := admin_controllers.NewFeatureController(db)
	usersController := admin_controllers.NewUserController(db)
	fgc := admin_controllers.NewFeatureGroupController(db)

	adminGroup := r.Group("/admin")
	adminGroup.Use(authentication.RequireRole("super_admin", db))

	adminGroup.GET("/package-tiers", packageTierController.GetAllTiers)
	adminGroup.GET("/package-tiers/:id", packageTierController.GetTier)
	adminGroup.POST("/package-tiers", packageTierController.CreateTier)
	adminGroup.PUT("/package-tiers/:id", packageTierController.UpdateTier)
	adminGroup.DELETE("/package-tiers/:id", packageTierController.DeleteTier)

	// Setting up routes
	adminGroup.GET("/plan-enrollments", PlanEnrollmentAdminController.GetAllEnrollments)
	adminGroup.GET("/plan-enrollments/:id", PlanEnrollmentAdminController.GetEnrollment)
	adminGroup.POST("/plan-enrollments", PlanEnrollmentAdminController.CreateEnrollment)
	adminGroup.PUT("/plan-enrollments/:id", PlanEnrollmentAdminController.UpdateEnrollment)
	adminGroup.DELETE("/plan-enrollments/:id", PlanEnrollmentAdminController.DeleteEnrollment)

	adminGroup.GET("/features", featureController.GetAllFeatures)
	adminGroup.GET("/features/:id", featureController.GetFeature)
	adminGroup.POST("/features", featureController.CreateFeature)
	adminGroup.PUT("/features/:id", featureController.UpdateFeature)
	adminGroup.DELETE("/features/:id", featureController.DeleteFeature)

	adminGroup.GET("/feature-groups", fgc.GetAllFeatureGroups)
	adminGroup.POST("/feature-groups", fgc.CreateFeatureGroup)
	adminGroup.GET("/feature-groups/:id", fgc.GetFeatureGroup)
	adminGroup.PUT("/feature-groups/:id", fgc.UpdateFeatureGroup)
	adminGroup.DELETE("/feature-groups/:id", fgc.DeleteFeatureGroup)

	// users
	adminGroup.GET("/users", usersController.ListUsers)
	adminGroup.GET("/users/:ug_kth_id", usersController.GetUser)
	adminGroup.PUT("/users/:ug_kth_id", usersController.UpdateUser)
	adminGroup.DELETE("/users/:ug_kth_id", usersController.DeleteUser)

	return r
}
