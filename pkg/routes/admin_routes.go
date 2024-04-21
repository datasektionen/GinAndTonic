package routes

import (
	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	admin_controllers "github.com/DowLucas/gin-ticket-release/pkg/controllers/admin"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminRoutes(r *gin.Engine, db *gorm.DB) *gin.Engine {
	pricingPackageAdminController := admin_controllers.NewPricingPackageAdminController(db)
	packageTierController := admin_controllers.NewPackageTierController(db)

	adminGroup := r.Group("/admin")
	adminGroup.Use(authentication.RequireRole("super_admin", db))

	adminGroup.GET("/package-tiers", packageTierController.GetAllTiers)
	adminGroup.GET("/package-tiers/:id", packageTierController.GetTier)
	adminGroup.POST("/package-tiers", packageTierController.CreateTier)
	adminGroup.PUT("/package-tiers/:id", packageTierController.UpdateTier)
	adminGroup.DELETE("/package-tiers/:id", packageTierController.DeleteTier)

	// Setting up routes
	adminGroup.GET("/pricing-packages", pricingPackageAdminController.GetAllPackages)
	adminGroup.GET("/pricing-packages/:id", pricingPackageAdminController.GetPackage)
	adminGroup.POST("/pricing-packages", pricingPackageAdminController.CreatePackage)
	adminGroup.PUT("/pricing-packages/:id", pricingPackageAdminController.UpdatePackage)
	adminGroup.DELETE("/pricing-packages/:id", pricingPackageAdminController.DeletePackage)

	return r
}
