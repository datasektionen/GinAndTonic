package routes

import (
	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers"
	manager_controller "github.com/DowLucas/gin-ticket-release/pkg/controllers/manager"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	feature_middleware "github.com/DowLucas/gin-ticket-release/pkg/middleware/feature"
	network_middlewares "github.com/DowLucas/gin-ticket-release/pkg/middleware/network"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ManagerRoutes(r *gin.Engine, db *gorm.DB) *gin.Engine {
	managerController := manager_controller.NewManagerController(db)
	eventController := controllers.NewEventController(db)
	eventWorkflowController := controllers.NewCompleteEventWorkflowController(db)
	ticketReleaseController := controllers.NewTicketReleaseController(db)
	ticketsController := controllers.NewTicketController(db)
	sendOutcontroller := controllers.NewSendOutController(db)
	salesReportController := controllers.NewSalesReportController(db)
	organizationController := controllers.NewOrganizationController(db)
	eventLandingPageController := controllers.NewEventLandingPageController(db)

	managerGroup := r.Group("/manager")
	managerGroup.Use(authentication.ValidateTokenMiddleware(true))
	managerGroup.Use(middleware.UserLoader(db))
	managerGroup.Use(middleware.RequireUserManager())

	managerGroup.GET("/network", managerController.GetNetworkDetails)

	// Events
	managerGroup.GET("/events", managerController.GetNetworkEvents)
	managerGroup.POST("/events", feature_middleware.RequireFeatureLimit(db, "max_events"), eventController.CreateEvent)
	managerGroup.POST("/complete-event-workflow", feature_middleware.RequireFeatureLimit(db, "max_events"), eventWorkflowController.CreateEvent)
	managerGroup.DELETE("/events/:eventID",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventController.DeleteEvent)
	managerGroup.POST("/events/:eventID/ticket-release",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		feature_middleware.SetParamObjectReference("eventID"),
		feature_middleware.RequireFeatureLimit(db, "max_ticket_releases_per_event"),
		ticketReleaseController.CreateTicketRelease)
	managerGroup.POST("/events/:eventID/tickets/qr-check-in",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		feature_middleware.RequireFeature(db, "check_in"),
		ticketsController.QrCodeCheckIn)

	// Event landing page
	managerGroup.GET("/events/:eventID/landing-page",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventLandingPageController.GetEventLandingPage)
	managerGroup.PUT("/events/:eventID/landing-page",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventLandingPageController.SaveEventLandingPage)
	managerGroup.GET("/events/:eventID/landing-page/editor",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventLandingPageController.GetEventLandingPageEditorState)
	managerGroup.PUT("/events/:eventID/landing-page/editor",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventLandingPageController.SaveEventLandingPageEditorState)
	managerGroup.PUT("/events/:eventID/landing-page/set-enabled",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventLandingPageController.ToggleLandingPageEnabled)

	// Sales report
	managerGroup.POST("/events/:eventID/sales-report",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		feature_middleware.RequireFeature(db, "sales_reports"),
		salesReportController.GenerateSalesReport)
	managerGroup.GET("/events/:eventID/sales-report",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		feature_middleware.RequireFeature(db, "sales_reports"),
		salesReportController.ListSalesReport)

	// Send outs
	managerGroup.GET("/events/:eventID/send-outs",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		feature_middleware.RequireFeature(db, "send_outs"),
		sendOutcontroller.GetEventSendOuts)
	managerGroup.POST("/events/:eventID/send-outs",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		feature_middleware.RequireFeature(db, "send_outs"),
		sendOutcontroller.SendOut)

	// Organizations
	managerGroup.POST("/organizations", network_middlewares.RequireNetworkRole(db, models.NetworkAdmin),
		feature_middleware.RequireFeatureLimit(db, "max_teams_per_network"),
		organizationController.CreateNetworkOrganization)

	return r
}
