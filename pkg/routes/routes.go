package routes

import (
	"net/http"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

var limiter = rate.NewLimiter(1, 5)

func rateLimitMiddleware(c *gin.Context) {
	if !limiter.Allow() {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		return
	}

	c.Next()
}

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	config := cors.DefaultConfig()
	env := os.Getenv("ENV")
	if env == "dev" {
		config.AllowOrigins = []string{"http://localhost:5000", "http://localhost", "http://localhost:8080"}
	} else if env == "prod" {
		config.AllowOrigins = []string{"https://tessera.datasektionen.se", "http://tessera.betasektionen.se"}
	}

	config.AllowCredentials = true

	r.Use(cors.New(config))

	r.GET("/postman-login", controllers.LoginPostman)
	r.GET("/postman-login-complete/:token", controllers.LoginCompletePostman)

	externalAuthService := services.NewExternalAuthService(db)
	externalAuthController := controllers.NewExternalAuthController(db, externalAuthService)

	r.POST("/external/signup", externalAuthController.SignupExternalUser)
	r.POST("/external/login", externalAuthController.LoginExternalUser)
	r.POST("/external/verify-email", externalAuthController.VerifyEmail)

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

	eventService := services.NewEventService(db)
	eventWorkflowController := controllers.NewCompleteEventWorkflowController(db, eventService)
	userController := controllers.NewUserController(db)

	organizationService := services.NewOrganizationService(db)
	allocateTicketsService := services.NewAllocateTicketsService(db)

	organizationController := controllers.NewOrganizationController(db, organizationService)
	ticketReleaseMethodsController := controllers.NewTicketReleaseMethodsController(db)
	ticketReleaseController := controllers.NewTicketReleaseController(db)
	ticketReleasePromoCodeController := controllers.NewTicketReleasePromoCodeController(db)
	ticketTypeController := controllers.NewTicketTypeController(db)
	organizationUsersController := controllers.NewOrganizationUsersController(db, organizationService)
	userFoodPreferenceController := controllers.NewUserFoodPreferenceController(db)
	ticketRequestController := controllers.NewTicketRequestController(db)
	allocateTicketsController := controllers.NewAllocateTicketsController(db, allocateTicketsService)
	ticketsController := controllers.NewTicketController(db)
	constantOptionsController := controllers.NewConstantOptionsController(db)
	paymentsController := controllers.NewPaymentController(db)
	notificationController := controllers.NewNotificationController(db)

	r.GET("/ticket-release/constants", constantOptionsController.ListTicketReleaseConstants)
	r.POST("/tickets/payment-webhook", paymentsController.PaymentWebhook)

	r.Use(authentication.ValidateTokenMiddleware())
	r.Use(middleware.UserLoader(db))

	//Event routes
	r.POST("/events", eventController.CreateEvent)
	r.GET("/events", eventController.ListEvents)
	r.GET("/events/:eventID", eventController.GetEvent)
	r.GET("/events/:eventID/manage", authentication.ValidateTokenMiddleware(),
		middleware.AuthorizeEventAccess(db), gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "User has access to this event"})
		}))

	r.PUT("/events/:eventID",
		authentication.RequireRole("user", db),
		eventController.UpdateEvent)
	r.DELETE("/events/:eventID",
		authentication.RequireRole("user", db),
		eventController.DeleteEvent)

	r.POST("/complete-event-workflow",
		authentication.RequireRole("user", db),
		eventWorkflowController.CreateEvent)

	// Ticket release routes
	r.GET("/events/:eventID/ticket-release", ticketReleaseController.ListEventTicketReleases)
	r.POST("/events/:eventID/ticket-release", middleware.AuthorizeEventAccess(db), ticketReleaseController.CreateTicketRelease)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db), ticketReleaseController.GetTicketRelease)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db), ticketReleaseController.UpdateTicketRelease)
	r.DELETE("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db), ticketReleaseController.DeleteTicketRelease)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/ticket-types", middleware.AuthorizeEventAccess(db), ticketTypeController.GetEventTicketTypes)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID/ticket-types", middleware.AuthorizeEventAccess(db), ticketTypeController.UpdateEventTicketTypes)
	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/manually-allocate-reserve-tickets",
		middleware.AuthorizeEventAccess(db),
		ticketReleaseController.ManuallyTryToAllocateReserveTickets)

	// Promo code routes
	r.GET("/activate-promo-code/:eventID", ticketReleasePromoCodeController.Create)

	// Allocate tickets routes
	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db), allocateTicketsController.AllocateTickets)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db), allocateTicketsController.ListAllocatedTickets)

	// Ticket request event routes
	r.GET("/events/:eventID/ticket-requests", ticketRequestController.Get)
	r.POST("/events/:eventID/ticket-requests", rateLimitMiddleware, ticketRequestController.Create)
	r.DELETE("/events/:eventID/ticket-requests/:ticketRequestID", ticketRequestController.CancelTicketRequest)

	// Ticket events routes
	r.GET("/events/:eventID/tickets", middleware.AuthorizeEventAccess(db), eventController.ListTickets)

	// My tickets
	r.GET("/my-ticket-requests", ticketRequestController.UsersList)
	r.GET("/my-tickets", ticketsController.UsersList)

	// Ticket routes
	r.DELETE("/my-tickets/:ticketID", ticketsController.CancelTicket)

	// Ticket routes
	r.GET("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db), ticketsController.GetTicket)
	r.PUT("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db), ticketsController.EditTicket)
	r.GET("/tickets/:ticketID/create-payment-intent", paymentsController.CreatePaymentIntent)

	r.POST("/organizations", authentication.RequireRole("super_admin", db), organizationController.CreateOrganization)
	r.GET("/organizations", authentication.RequireRole("super_admin", db), organizationController.ListOrganizations)
	r.GET("my-organizations", organizationController.ListMyOrganizations)
	r.GET("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.GetOrganization)
	r.PUT("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.UpdateOrganization)
	r.DELETE("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db), organizationController.DeleteOrganization)
	r.GET("/organizations/:organizationID/events", middleware.AuthorizeOrganizationAccess(db), organizationController.ListOrganizationEvents)

	// Organization Users routes
	r.GET("/organizations/:organizationID/users", middleware.AuthorizeOrganizationRole(db, models.OrganizationMember), organizationUsersController.GetOrganizationUsers)
	r.POST("/organizations/:organizationID/users/:username", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.AddUserToOrganization)
	r.DELETE("/organizations/:organizationID/users/:username", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.RemoveUserFromOrganization)
	r.PUT("/organizations/:organizationID/users/:username", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.ChangeUserOrganizationRole)

	// Ticket Release Methods routes
	r.GET("/ticket-release-methods", ticketReleaseMethodsController.ListTicketReleaseMethods)
	r.POST("/ticket-release-methods", authentication.RequireRole("super_admin", db), ticketReleaseMethodsController.CreateTicketReleaseMethod)

	// Ticket Types routes
	r.GET("/ticket-types",
		authentication.RequireRole("super_admin", db),
		ticketTypeController.ListAllTicketTypes)
	r.POST("/ticket-types", authentication.RequireRole("super_admin", db),
		ticketTypeController.CreateTicketTypes)

	// User Food Preference routes
	r.PUT("/user-food-preferences", userFoodPreferenceController.Update)
	r.GET("/user-food-preferences", userFoodPreferenceController.Get)
	r.GET("/food-preferences", userFoodPreferenceController.ListFoodPreferences)

	r.POST("/admin/create-user", authentication.RequireRole("super_admin", db), userController.CreateUser)

	// Testing
	r.POST("send-test-email", authentication.RequireRole("super_admin", db), notificationController.SendTestEmail)
	return r
}
