package routes

import (
	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers"
	manager_controller "github.com/DowLucas/gin-ticket-release/pkg/controllers/manager"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	feature_middleware "github.com/DowLucas/gin-ticket-release/pkg/middleware/feature"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ManagerRoutes(r *gin.Engine, db *gorm.DB) *gin.Engine {
	managerController := manager_controller.NewManagerController(db)
	eventController := controllers.NewEventController(db)
	eventWorkflowController := controllers.NewCompleteEventWorkflowController(db)
	ticketReleaseController := controllers.NewTicketReleaseController(db)

	managerGroup := r.Group("/manager")
	managerGroup.Use(authentication.ValidateTokenMiddleware(true))
	managerGroup.Use(middleware.UserLoader(db))
	managerGroup.Use(middleware.RequireUserManager())

	managerGroup.GET("/events", managerController.GetNetworkEvents)
	managerGroup.POST("/events", feature_middleware.RequireFeatureLimit(db, "max_events"), eventController.CreateEvent)
	managerGroup.POST("/complete-event-workflow", feature_middleware.RequireFeatureLimit(db, "max_events"), eventWorkflowController.CreateEvent)
	managerGroup.POST("/events/:eventID/ticket-release",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		feature_middleware.SetParamObjectReference("eventID"),
		feature_middleware.RequireFeatureLimit(db, "max_ticket_releases"),
		ticketReleaseController.CreateTicketRelease)

	managerGroup.GET("/network", managerController.GetNetworkDetails)

	return r
}
