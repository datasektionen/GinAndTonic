package routes

import (
	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5000"}
	config.AllowCredentials = true

	r.Use(cors.New(config))

	r.GET("/postman-login", controllers.LoginPostman)
	r.GET("/postman-login-complete/:token", controllers.LoginCompletePostman)

	r.GET("/login", controllers.Login)
	r.GET("/login-complete/:token", controllers.LoginComplete)
	r.GET("/current-user", authentication.ValidateTokenMiddleware(), controllers.CurrentUser)
	r.GET("/logout", authentication.ValidateTokenMiddleware(), controllers.Logout)
	r.GET("/event", gin.HandlerFunc(func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}))
	eventController := controllers.NewEventController(db)

	organizationService := services.NewOrganizationService(db)
	allocateTicketsService := services.NewAllocateTicketsService(db)

	organizationController := controllers.NewOrganizationController(db, organizationService)
	ticketReleaseMethodsController := controllers.NewTicketReleaseMethodsController(db)
	ticketReleaseController := controllers.NewTicketReleaseController(db)
	ticketTypeController := controllers.NewTicketTypeController(db)
	organizationUsersController := controllers.NewOrganizationUsersController(db, organizationService)
	userFoodPreferenceController := controllers.NewUserFoodPreferenceController(db)
	ticketRequestController := controllers.NewTicketRequestController(db)
	allocateTicketsController := controllers.NewAllocateTicketsController(db, allocateTicketsService)
	ticketsController := controllers.NewTicketController(db)

	constantOptionsController := controllers.NewConstantOptionsController(db)
	r.GET("/ticket-release/constants", constantOptionsController.ListTicketReleaseConstants)

	r.Use(authentication.ValidateTokenMiddleware())

	//Event routes
	r.POST("/events", eventController.CreateEvent)
	r.GET("/events", eventController.ListEvents)
	r.GET("/events/:eventID", middleware.AuthorizeEventAccess(db), eventController.GetEvent)
	r.PUT("/events/:eventID", middleware.AuthorizeEventAccess(db), eventController.UpdateEvent)
	r.DELETE("/events/:eventID", middleware.AuthorizeEventAccess(db), eventController.DeleteEvent)

	// Ticket release routes
	r.GET("/events/:eventID/ticket-release", ticketReleaseController.ListEventTicketReleases)
	r.POST("/events/:eventID/ticket-release", ticketReleaseController.CreateTicketRelease)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID", ticketReleaseController.GetTicketRelease)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db), ticketReleaseController.UpdateTicketRelease)
	r.DELETE("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db), ticketReleaseController.DeleteTicketRelease)

	// Allocate tickets routes
	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db), allocateTicketsController.AllocateTickets)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db), allocateTicketsController.ListAllocatedTickets)

	// Ticket request routes
	r.GET("/events/:eventID/ticket-requests", ticketRequestController.Get)
	r.POST("/events/:eventID/ticket-requests", ticketRequestController.Create)

	// Ticket routes
	r.GET("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db), ticketsController.GetTicket)
	r.PUT("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db), ticketsController.EditTicket)

	r.POST("/organizations", organizationController.CreateOrganization)
	r.GET("/organizations", organizationController.ListOrganizations)
	r.GET("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.GetOrganization)
	r.PUT("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.UpdateOrganization)
	r.DELETE("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.DeleteOrganization)

	// Organization Users routes
	r.GET("/organizations/:organizationID/users", middleware.AuthorizeOrganizationRole(db, models.OrganizationMember), organizationUsersController.GetOrganizationUsers)
	r.POST("/organizations/:organizationID/users/:ugkthid", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.AddUserToOrganization)
	r.DELETE("/organizations/:organizationID/users/:ugkthid", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.RemoveUserFromOrganization)

	// Ticket Release Methods routes
	r.GET("/ticket-release-methods", ticketReleaseMethodsController.ListTicketReleaseMethods)
	r.POST("/ticket-release-methods", authentication.RequireRole("super_admin"), ticketReleaseMethodsController.CreateTicketReleaseMethod)

	// Ticket Types routes
	r.GET("/ticket-types", authentication.ValidateTokenMiddleware(), ticketTypeController.ListAllTicketTypes)
	r.POST("/ticket-types", authentication.ValidateTokenMiddleware(), ticketTypeController.CreateTicketTypes)

	// User Food Preference routes
	r.PUT("/user-food-preferences", userFoodPreferenceController.Update)
	r.GET("/user-food-preferences", userFoodPreferenceController.Get)

	return r
}
