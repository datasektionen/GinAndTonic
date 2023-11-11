package routes

import (
	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/login", controllers.Login)
	r.GET("/login-complete/:token", controllers.LoginComplete)
	r.GET("/logout", authentication.ValidateTokenMiddleware(), controllers.Logout)

	organizationService := services.NewOrganizationService(db)
	allocateTicketsService := services.NewAllocateTicketsService(db)

	eventController := controllers.NewEventController(db)
	organizationController := controllers.NewOrganizationController(db, organizationService)
	ticketReleaseMethodsController := controllers.NewTicketReleaseMethodsController(db)
	ticketReleaseController := controllers.NewTicketReleaseController(db)
	ticketTypeController := controllers.NewTicketTypeController(db)
	organizationUsersController := controllers.NewOrganizationUsersController(db, organizationService)
	userFoodPreferenceController := controllers.NewUserFoodPreferenceController(db)
	ticketRequestController := controllers.NewTicketRequestController(db)
	allocateTicketsController := controllers.NewAllocateTicketsController(db, allocateTicketsService)

	// Group event-related routes together
	eventGroup := r.Group("/events")
	{
		eventGroup.Use(authentication.ValidateTokenMiddleware())
		eventGroup.GET("/", eventController.ListEvents)                                                  // List all events
		eventGroup.POST("/", eventController.CreateEvent)                                                // Create a new event
		eventGroup.GET("/:eventID", middleware.AuthorizeEventAccess(db), eventController.GetEvent)       // Get an event by ID
		eventGroup.PUT("/:eventID", middleware.AuthorizeEventAccess(db), eventController.UpdateEvent)    // Update an event by ID
		eventGroup.DELETE("/:eventID", middleware.AuthorizeEventAccess(db), eventController.DeleteEvent) // Delete an event by ID

		ticketRelease := eventGroup.Group("/:eventID/ticket-release")
		{
			ticketRelease.Use(authentication.ValidateTokenMiddleware())
			ticketRelease.GET("/", ticketReleaseController.ListEventTicketReleases)
			ticketRelease.POST("/", ticketReleaseController.CreateTicketRelease)
			ticketRelease.GET("/:ticketReleaseID", ticketReleaseController.GetTicketRelease)
			ticketRelease.PUT("/:ticketReleaseID", middleware.AuthorizeEventAccess(db), ticketReleaseController.UpdateTicketRelease)
			ticketRelease.DELETE("/:ticketReleaseID", middleware.AuthorizeEventAccess(db), ticketReleaseController.DeleteTicketRelease)

			allocateTickets := ticketRelease.Group("/:ticketReleaseID/allocate-tickets")
			{
				allocateTickets.Use(authentication.ValidateTokenMiddleware())
				allocateTickets.POST("/", middleware.AuthorizeEventAccess(db), allocateTicketsController.AllocateTickets)
				allocateTickets.GET("/", middleware.AuthorizeEventAccess(db), allocateTicketsController.ListAllocatedTickets)
			}
		}

		ticketRequests := eventGroup.Group("/:eventID/ticket-requests")
		{
			ticketRequests.Use(authentication.ValidateTokenMiddleware())
			ticketRequests.GET("/", ticketRequestController.Get)
			ticketRequests.POST("/", ticketRequestController.Create)
		}
	}

	organizations := r.Group("/organizations")
	{
		organizations.Use(authentication.ValidateTokenMiddleware())
		organizations.POST("", organizationController.CreateOrganization)
		organizations.GET("", organizationController.ListOrganizations)
		organizations.GET("/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.GetOrganization)
		organizations.PUT("/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.UpdateOrganization)
		organizations.DELETE("/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.DeleteOrganization)

		organizationUsers := organizations.Group("/:organizationID/users")
		{
			organizationUsers.GET("/", middleware.AuthorizeOrganizationRole(db, models.OrganizationMember), organizationUsersController.GetOrganizationUsers)
			organizationUsers.POST("/:ugkthid", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.AddUserToOrganization)
			organizationUsers.DELETE("/:ugkthid", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.RemoveUserFromOrganization)
		}
	}

	ticketReleaseMethods := r.Group("/ticket-release-methods")
	{
		ticketReleaseMethods.Use(authentication.ValidateTokenMiddleware())
		ticketReleaseMethods.GET("/", ticketReleaseMethodsController.ListTicketReleaseMethods)
		ticketReleaseMethods.POST("/", authentication.RequireRole("super_admin"), ticketReleaseMethodsController.CreateTicketReleaseMethod)
	}

	ticketTypes := r.Group("/ticket-types")
	{
		ticketTypes.GET("/", authentication.ValidateTokenMiddleware(), ticketTypeController.ListAllTicketTypes)
		ticketTypes.POST("/", authentication.ValidateTokenMiddleware(), ticketTypeController.CreateTicketTypes)
	}

	userFoodPreference := r.Group("/user-food-preferences")
	{
		userFoodPreference.Use(authentication.ValidateTokenMiddleware())
		userFoodPreference.PUT("/", userFoodPreferenceController.Update)
		userFoodPreference.GET("/", userFoodPreferenceController.Get)
	}

	return r
}
