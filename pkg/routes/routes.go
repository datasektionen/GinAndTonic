package routes

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers"
	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	banking_service "github.com/DowLucas/gin-ticket-release/pkg/services/banking"
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

	// Legal
	r.Static("/static", "./static")

	r.GET("/privacy-policy", func(c *gin.Context) {
		c.File("./static/privacy.html")
	})

	r.GET("/postman-login", controllers.LoginPostman)
	r.GET("/postman-login-complete/:token", controllers.LoginCompletePostman)

	externalAuthService := services.NewExternalAuthService(db)
	externalAuthController := controllers.NewExternalAuthController(db, externalAuthService)
	passwordResetController := controllers.NewUserPasswordResetController(db)

	r.POST("/external/signup", externalAuthController.SignupExternalUser)
	r.POST("/external/login", externalAuthController.LoginExternalUser)
	r.POST("/external/verify-email", externalAuthController.VerifyEmail)
	r.POST("/external/resend-verification-email", externalAuthController.ResendVerificationEmail)

	// Password reset
	r.POST("/password-reset", passwordResetController.CreatePasswordReset)
	r.POST("/password-reset/complete", passwordResetController.CompletePasswordReset)

	r.GET("/login", controllers.Login)
	r.GET("/login-complete/:token", controllers.LoginComplete)
	r.GET("/current-user", authentication.ValidateTokenMiddleware(), controllers.CurrentUser)
	r.GET("/logout", authentication.ValidateTokenMiddleware(), controllers.Logout)

	eventController := controllers.NewEventController(db)

	completeEventWorkflowService := services.NewCompleteEventWorkflowService(db)
	eventWorkflowController := controllers.NewCompleteEventWorkflowController(db, completeEventWorkflowService)
	userController := controllers.NewUserController(db)
	sendOutService := services.NewSendOutService(db)
	teamService := services.NewTeamService(db)
	allocateTicketsService := services.NewAllocateTicketsService(db)
	preferredEmailService := services.NewPreferredEmailService(db)
	bankingService := banking_service.NewBankingService(db)

	teamController := controllers.NewTeamController(db, teamService)
	ticketReleaseMethodsController := controllers.NewTicketReleaseMethodsController(db)
	ticketReleaseController := controllers.NewTicketReleaseController(db)
	ticketReleasePromoCodeController := controllers.NewTicketReleasePromoCodeController(db)
	ticketTypeController := controllers.NewTicketTypeController(db)
	teamUsersController := controllers.NewTeamUsersController(db, teamService)
	userFoodPreferenceController := controllers.NewUserFoodPreferenceController(db)
	ticketRequestController := controllers.NewTicketRequestController(db)
	allocateTicketsController := controllers.NewAllocateTicketsController(db, allocateTicketsService)
	ticketsController := controllers.NewTicketController(db)
	constantOptionsController := controllers.NewConstantOptionsController(db)
	paymentsController := controllers.NewPaymentController(db)
	notificationController := controllers.NewNotificationController(db)
	contactController := controllers.NewContactController(db)
	ticketReleaseReminderController := controllers.NewTicketReleaseReminderController(db)
	sendOutcontroller := controllers.NewSendOutController(db, sendOutService)
	salesReportController := controllers.NewSalesReportController(db)
	preferredEmailController := controllers.NewPreferredEmailController(db, preferredEmailService)
	eventFormFieldController := controllers.NewEventFormFieldController(db)
	eventFromFieldResponseController := controllers.NewEventFormFieldResponseController(db)
	addOnController := controllers.NewAddOnController(db)
	eventSiteVistsController := controllers.NewSitVisitsController(db)
	bankingController := controllers.NewBankingController(bankingService)

	r.GET("/ticket-release/constants", constantOptionsController.ListTicketReleaseConstants)
	r.POST("/tickets/payment-webhook", paymentsController.PaymentWebhook)

	r.POST("/preferred-email/verify", preferredEmailController.Verify)

	r.Use(authentication.ValidateTokenMiddleware())
	r.Use(middleware.UserLoader(db))

	synqMonHandler := setupAsynqMon()
	r.Any("/admin/monitoring/*any", authentication.RequireRole("super_admin", db), gin.WrapH(synqMonHandler)) // Serve asynqmon on /monitoring path

	//Event routes
	r.POST("/events", eventController.CreateEvent)
	r.GET("/events", eventController.ListEvents)
	r.GET("/events/:eventID", middleware.UpdateSiteVisits(db), eventController.GetEvent)
	r.GET("/events/:eventID/manage", authentication.ValidateTokenMiddleware(),
		middleware.AuthorizeEventAccess(db, models.TeamMember), gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "User has access to this event"})
		}))

	r.GET("/test", authentication.RequireRole("super_admin", db), func(c *gin.Context) {
		jobs.StartEventSiteVisitsJob(db)
		c.JSON(http.StatusOK, gin.H{"message": "Job started"})
	})

	r.GET("/events/:eventID/manage/secret-token",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		eventController.GetEventSecretToken)

	r.PUT("/events/:eventID",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		eventController.UpdateEvent)
	r.DELETE("/events/:eventID",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		eventController.DeleteEvent)

	r.POST("/complete-event-workflow",
		eventWorkflowController.CreateEvent)

	// Contact
	r.POST("/contact", contactController.CreateContact)

	// Ticket release routes
	r.GET("/events/:eventID/ticket-release", ticketReleaseController.ListEventTicketReleases)
	r.POST("/events/:eventID/ticket-release", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketReleaseController.CreateTicketRelease)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketReleaseController.GetTicketRelease)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketReleaseController.UpdateTicketRelease)
	r.DELETE("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketReleaseController.DeleteTicketRelease)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/ticket-types", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketTypeController.GetEventTicketTypes)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID/ticket-types", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketTypeController.UpdateEventTicketTypes)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID/payment-deadline",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		ticketReleaseController.UpdatePaymentDeadline)

	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/manually-allocate-reserve-tickets",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		ticketReleaseController.ManuallyTryToAllocateReserveTickets)

	// Site vists
	r.GET("/events/:eventID/overview", middleware.AuthorizeEventAccess(db, models.TeamMember), eventSiteVistsController.Get)

	// AddOn
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/add-ons",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		addOnController.GetAddOns)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID/add-ons",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		addOnController.UpsertAddOns)
	r.DELETE("/events/:eventID/ticket-release/:ticketReleaseID/add-ons/:addOnID",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		addOnController.DeleteAddOn)

	// Form fields
	r.PUT("/events/:eventID/form-fields", eventFormFieldController.Upsert)
	r.PUT("/events/:eventID/ticket-requests/:ticketRequestID/form-fields", eventFromFieldResponseController.Upsert)

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
	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db, models.TeamMember), allocateTicketsController.AllocateTickets)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db, models.TeamMember), allocateTicketsController.ListAllocatedTickets)
	r.POST("/events/:eventID/ticket-requests/:ticketRequestID/allocate",
		middleware.AuthorizeEventAccess(db, models.TeamMember),
		allocateTicketsController.SelectivelyAllocateTicketRequest)

	rlm := NewRateLimiterMiddleware(2, 5) // For example, 1 request per second with a burst of 5

	// Ticket request event routes
	r.GET("/events/:eventID/ticket-requests", ticketRequestController.Get)
	r.POST("/events/:eventID/ticket-requests", rlm.MiddlewareFunc(), ticketRequestController.Create)
	r.DELETE("/events/:eventID/ticket-requests/:ticketRequestID", ticketRequestController.CancelTicketRequest)
	r.PUT("/ticket-releases/:ticketReleaseID/ticket-requests/:ticketRequestID/add-ons", ticketRequestController.UpdateAddOns)

	// Ticket events routes
	r.GET("/events/:eventID/tickets", middleware.AuthorizeEventAccess(db, models.TeamMember), eventController.ListTickets)
	r.POST("/events/:eventID/tickets/qr-check-in", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketsController.QrCodeCheckIn)

	// My tickets
	r.GET("/my-ticket-requests", ticketRequestController.UsersList)
	r.GET("/my-tickets", ticketsController.UsersList)

	// Ticket routes
	r.DELETE("/my-tickets/:ticketID", ticketsController.CancelTicket)

	// send outs
	r.GET("/events/:eventID/send-outs", middleware.AuthorizeEventAccess(db, models.TeamMember), sendOutcontroller.GetEventSendOuts)
	r.POST("/events/:eventID/send-out", middleware.AuthorizeEventAccess(db, models.TeamMember), sendOutcontroller.SendOut)

	// Ticket routes
	r.GET("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketsController.GetTicket)
	r.PUT("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db, models.TeamMember), ticketsController.UpdateTicket)
	r.GET("/tickets/:ticketID/create-payment-intent", paymentsController.CreatePaymentIntent)

	// Sales report
	r.POST("/events/:eventID/sales-report", middleware.AuthorizeEventAccess(db, models.TeamMember), salesReportController.GenerateSalesReport)
	r.GET("/events/:eventID/sales-report", middleware.AuthorizeEventAccess(db, models.TeamMember), salesReportController.ListSalesReport)

	r.POST("/teams", authentication.RequireRole("super_admin", db), teamController.CreateTeam)
	r.GET("/teams", teamController.ListTeams)
	r.GET("my-teams", teamController.ListMyTeams)
	r.GET("/teams/:teamID", middleware.AuthorizeTeamAccess(db, models.TeamMember), teamController.GetTeam)
	r.PUT("/teams/:teamID", middleware.AuthorizeTeamAccess(db, models.TeamMember), teamController.UpdateTeam)
	r.DELETE("/teams/:teamID", middleware.AuthorizeTeamAccess(db, models.TeamMember), teamController.DeleteTeam)
	r.GET("/teams/:teamID/events", middleware.AuthorizeTeamAccess(db, models.TeamMember), teamController.ListTeamEvents)

	// Team Users routes
	r.GET("/teams/:teamID/users", middleware.AuthorizeTeamRole(db, models.TeamMember), teamUsersController.GetTeamUsers)
	r.POST("/teams/:teamID/users/:username", middleware.AuthorizeTeamRole(db, models.TeamOwner), teamUsersController.AddUserToTeam)
	r.DELETE("/teams/:teamID/users/:username", middleware.AuthorizeTeamRole(db, models.TeamOwner), teamUsersController.RemoveUserFromTeam)
	r.PUT("/teams/:teamID/users/:username", middleware.AuthorizeTeamRole(db, models.TeamOwner), teamUsersController.ChangeUserTeamRole)

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

	// Banking details
	r.GET("/teams/:teamID/banking-details", middleware.AuthorizeTeamRole(db, models.TeamMember), bankingController.GetBankingDetails)
	r.POST("/teams/:teamID/banking-details", middleware.AuthorizeTeamRole(db, models.TeamOwner), bankingController.SubmitBankingDetails)
	r.DELETE("/teams/:teamID/banking-details", middleware.AuthorizeTeamRole(db, models.TeamOwner), bankingController.DeleteBankingDetails)

	// Preferred email
	r.POST("/preferred-email/request", preferredEmailController.Request)

	r.POST("send-test-email", authentication.RequireRole("super_admin", db), notificationController.SendTestEmail)
	return r
}
