package routes

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"gorm.io/gorm"
)

func setupAsynqMon() *asynqmon.HTTPHandler {
	// Parse the REDIS_URL
	redisURL, err := url.Parse(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("Failed to parse REDIS_URL: %v", err)
	}

	// Extract host and password. Port is part of the host in the URL.
	redisHost := redisURL.Host
	redisPassword, _ := redisURL.User.Password()

	// Setup asynqmon based on the environment
	var h *asynqmon.HTTPHandler
	if os.Getenv("ENV") == "dev" {
		h = asynqmon.New(asynqmon.Options{
			RootPath:     "/admin/monitoring",
			RedisConnOpt: asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")},
		})
	} else {
		h = asynqmon.New(asynqmon.Options{
			RootPath: "/admin/monitoring",
			RedisConnOpt: asynq.RedisClientOpt{
				Addr:     redisHost,
				Password: redisPassword,
			},
		})
	}

	return h
}

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	config := cors.DefaultConfig()
	if os.Getenv("ENV") == "dev" {
		config.AllowOrigins = []string{"http://localhost:5000", "http://localhost", "http://localhost:8080"}
	} else if os.Getenv("ENV") == "prod" {
		config.AllowOrigins = []string{"https://tessera.datasektionen.se"}
	}

	config.AllowCredentials = true

	r.Use(cors.New(config))

	r.GET("/postman-login", controllers.LoginPostman)
	r.GET("/postman-login-complete/:token", controllers.LoginCompletePostman)

	externalAuthService := services.NewExternalAuthService(db)
	externalAuthController := controllers.NewExternalAuthController(db, externalAuthService)
	passwordResetController := controllers.NewUserPasswordResetController(db)

	r.POST("/external/signup", externalAuthController.SignupExternalUser)
	r.POST("/external/login", externalAuthController.LoginExternalUser)
	r.POST("/external/verify-email", externalAuthController.VerifyEmail)

	// Password reset
	r.POST("/password-reset", passwordResetController.CreatePasswordReset)
	r.POST("/password-reset/complete", passwordResetController.CompletePasswordReset)

	r.GET("/login", controllers.Login)
	r.GET("/login-complete/:token", controllers.LoginComplete)
	r.GET("/current-user", authentication.ValidateTokenMiddleware(), controllers.CurrentUser)
	r.GET("/logout", authentication.ValidateTokenMiddleware(), controllers.Logout)

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
	contactController := controllers.NewContactController(db)
	ticketReleaseReminderController := controllers.NewTicketReleaseReminderController(db)

	r.GET("/ticket-release/constants", constantOptionsController.ListTicketReleaseConstants)
	r.POST("/tickets/payment-webhook", paymentsController.PaymentWebhook)

	r.Use(authentication.ValidateTokenMiddleware())
	r.Use(middleware.UserLoader(db))

	synqMonHandler := setupAsynqMon()
	r.GET("/admin/monitoring/*any", authentication.RequireRole("super_admin", db), gin.WrapH(synqMonHandler)) // Serve asynqmon on /monitoring path

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

	// Contact
	r.POST("/contact", contactController.CreateContact)

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

	// Ticket release reminder
	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/reminder",
		ticketReleaseReminderController.CreateTicketReleaseReminder)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/reminder",
		ticketReleaseReminderController.GetTicketReleaseReminder)
	r.DELETE("/events/:eventID/ticket-release/:ticketReleaseID/reminder",
		ticketReleaseReminderController.DeleteTicketReleaseReminder)

	// Promo code routes
	r.GET("/activate-promo-code/:eventID", ticketReleasePromoCodeController.Create)

	// Allocate tickets routes
	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db), allocateTicketsController.AllocateTickets)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db), allocateTicketsController.ListAllocatedTickets)

	rlm := NewRateLimiterMiddleware(2, 5) // For example, 1 request per second with a burst of 5

	// Ticket request event routes
	r.GET("/events/:eventID/ticket-requests", ticketRequestController.Get)
	r.POST("/events/:eventID/ticket-requests", rlm.MiddlewareFunc(), ticketRequestController.Create)
	r.DELETE("/events/:eventID/ticket-requests/:ticketRequestID", ticketRequestController.CancelTicketRequest)

	// Ticket events routes
	r.GET("/events/:eventID/tickets", middleware.AuthorizeEventAccess(db), eventController.ListTickets)
	r.POST("/events/:eventID/tickets/qr-check-in", middleware.AuthorizeEventAccess(db), ticketsController.QrCodeCheckIn)

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
