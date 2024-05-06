package routes

import (
	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	manager_controller "github.com/DowLucas/gin-ticket-release/pkg/controllers/manager"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ManagerRoutes(r *gin.Engine, db *gorm.DB) *gin.Engine {

	managerController := manager_controller.NewManagerController(db)

	managerGroup := r.Group("/manager")
	managerGroup.Use(authentication.ValidateTokenMiddleware(true))
	managerGroup.Use(middleware.UserLoader(db))
	managerGroup.Use(middleware.RequireUserManager())

	managerGroup.GET("/events", managerController.GetNetworkEvents)
	managerGroup.GET("/network", managerController.GetNetworkDetails)

	return r
}
