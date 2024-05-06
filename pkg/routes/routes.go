package routes

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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
	"golang.org/x/time/rate"
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
	// config := cors.DefaultConfig()

	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Range"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept", "Authorization", "Range"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

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

	// Deprecated routes, to be removed
	// r.GET("/postman-login", controllers.LoginPostman)
	// r.GET("/postman-login-complete/:token", controllers.LoginCompletePostman)

	customerAuthService := services.NewCustomerAuthService(db)
	customerAuthController := controllers.NewCustomerAuthController(db, customerAuthService)
	passwordResetController := controllers.NewUserPasswordResetController(db)

	r.POST("/customer/signup", customerAuthController.SignupCustomerUser)
	r.POST("/customer/login", customerAuthController.LoginUser)
	r.POST("/customer/verify-email", customerAuthController.VerifyEmail)
	r.POST("/customer/resend-verification-email", customerAuthController.ResendVerificationEmail)

	// Password reset
	r.POST("/password-reset", passwordResetController.CreatePasswordReset)
	r.POST("/password-reset/complete", passwordResetController.CompletePasswordReset)

	// Deprecated routes, to be removed
	// r.GET("/login", controllers.Login)
	// r.GET("/login-complete/:token", controllers.LoginComplete)

	r.GET("/current-user", authentication.ValidateTokenMiddleware(true), controllers.CurrentUser)
	r.GET("/logout", authentication.ValidateTokenMiddleware(true), controllers.Logout)

	eventController := controllers.NewEventController(db)

	completeEventWorkflowService := services.NewCompleteEventWorkflowService(db)
	eventWorkflowController := controllers.NewCompleteEventWorkflowController(db, completeEventWorkflowService)
	userController := controllers.NewUserController(db)
	sendOutService := services.NewSendOutService(db)
	organizationService := services.NewOrganizationService(db)
	allocateTicketsService := services.NewAllocateTicketsService(db)
	bankingService := banking_service.NewBankingService(db)

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
	sendOutcontroller := controllers.NewSendOutController(db, sendOutService)
	salesReportController := controllers.NewSalesReportController(db)
	eventFormFieldController := controllers.NewEventFormFieldController(db)
	eventFromFieldResponseController := controllers.NewEventFormFieldResponseController(db)
	addOnController := controllers.NewAddOnController(db)
	eventSiteVistsController := controllers.NewSitVisitsController(db)
	bankingController := controllers.NewBankingController(bankingService)
	guestController := controllers.NewGuestController(db)

	var rlm *RateLimiterMiddleware
	var rlmURLParam *RateLimiterMiddleware
	if os.Getenv("ENV") == "dev" {
		// For development, we dont really care about the rate limit
		rlm = NewRateLimiterMiddleware(2, 5)
		rlmURLParam = NewRateLimiterMiddleware(2, 5)
	} else {
		rlm = NewRateLimiterMiddleware(rate.Limit(1.0/60.0), 1)
		rlmURLParam = NewRateLimiterMiddleware(rate.Limit(1.0/60.0), 1)
	}

	r.GET("/ticket-release/constants", constantOptionsController.ListTicketReleaseConstants)
	r.POST("/tickets/payment-webhook", paymentsController.PaymentWebhook)

	r.GET("/view/events/:refID", authentication.ValidateTokenMiddleware(false), middleware.UpdateSiteVisits(db), eventController.GetEvent)

	r.GET("/guest-customer/:ugkthid/activate-promo-code/:eventID", ticketReleasePromoCodeController.GuestCreate)
	r.GET("/guest-customer/:ugkthid/tickets/:ticketID/create-payment-intent", paymentsController.GuestCreatePaymentIntent)
	r.GET("/guest-customer/:ugkthid", guestController.Get)
	r.DELETE("/guest-customer/:ugkthid/ticket-requests/:ticketRequestID", ticketRequestController.GuestCancelTicketRequest)
	r.DELETE("/guest-customer/:ugkthid/my-tickets/:ticketID", ticketsController.GuestCancelTicket)
	r.PUT("/guest-customer/:ugkthid/events/:eventID/ticket-requests/:ticketRequestID/form-fields", eventFromFieldResponseController.GuestUpsert)
	r.GET("/guest-customer/:ugkthid/user-food-preferences", userFoodPreferenceController.GuestGet)
	r.PUT("/guest-customer/:ugkthid/user-food-preferences", userFoodPreferenceController.GuestUpdate)
	r.POST("/guest-customer/:ugkthid/events/:eventID/guest-customer/ticket-requests", rlmURLParam.MiddlewareFuncURLParam(), ticketRequestController.GuestCreate)

	r.Use(authentication.ValidateTokenMiddleware(true))
	r.Use(middleware.UserLoader(db))

	synqMonHandler := setupAsynqMon()
	r.Any("/admin/monitoring/*any", authentication.RequireRole(models.RoleSuperAdmin, db), gin.WrapH(synqMonHandler)) // Serve asynqmon on /monitoring path

	//Event routes
	r.POST("/events", eventController.CreateEvent)
	r.GET("/events", eventController.ListEvents)
	r.GET("/events/:eventID", middleware.UpdateSiteVisits(db), eventController.GetEvent)
	r.GET("/events/:eventID/manage",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember), gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "User has access to this event"})
		}))

	r.GET("/test", authentication.RequireRole(models.RoleSuperAdmin, db), func(c *gin.Context) {
		jobs.StartEventSiteVisitsJob(db)
		c.JSON(http.StatusOK, gin.H{"message": "Job started"})
	})

	r.GET("/events/:eventID/manage/secret-token",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventController.GetEventSecretToken)

	r.PUT("/events/:eventID",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventController.UpdateEvent)
	r.DELETE("/events/:eventID",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventController.DeleteEvent)

	r.POST("/complete-event-workflow",
		eventWorkflowController.CreateEvent)

	// Contact
	r.POST("/contact", contactController.CreateContact)
	r.POST("/plan-contact", contactController.CreatePlanContact)

	// Ticket release routes
	r.GET("/events/:eventID/ticket-release", ticketReleaseController.ListEventTicketReleases)
	r.POST("/events/:eventID/ticket-release", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketReleaseController.CreateTicketRelease)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketReleaseController.GetTicketRelease)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketReleaseController.UpdateTicketRelease)
	r.DELETE("/events/:eventID/ticket-release/:ticketReleaseID", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketReleaseController.DeleteTicketRelease)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/ticket-types", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketTypeController.GetEventTicketTypes)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID/ticket-types", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketTypeController.UpdateEventTicketTypes)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID/payment-deadline",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		ticketReleaseController.UpdatePaymentDeadline)

	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/manually-allocate-reserve-tickets",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		ticketReleaseController.ManuallyTryToAllocateReserveTickets)

	// Site vists
	r.GET("/events/:eventID/overview", middleware.AuthorizeEventAccess(db, models.OrganizationMember), eventSiteVistsController.Get)

	// AddOn
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/add-ons",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		addOnController.GetAddOns)
	r.PUT("/events/:eventID/ticket-release/:ticketReleaseID/add-ons",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		addOnController.UpsertAddOns)
	r.DELETE("/events/:eventID/ticket-release/:ticketReleaseID/add-ons/:addOnID",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
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
	r.POST("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db, models.OrganizationMember), allocateTicketsController.AllocateTickets)
	r.GET("/events/:eventID/ticket-release/:ticketReleaseID/allocate-tickets", middleware.AuthorizeEventAccess(db, models.OrganizationMember), allocateTicketsController.ListAllocatedTickets)
	r.POST("/events/:eventID/ticket-requests/:ticketRequestID/allocate",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		allocateTicketsController.SelectivelyAllocateTicketRequest)

	// Ticket request event routes
	r.GET("/events/:eventID/ticket-requests", ticketRequestController.Get)
	r.POST("/events/:eventID/ticket-requests", rlm.MiddlewareFunc(), ticketRequestController.Create)
	r.DELETE("/events/:eventID/ticket-requests/:ticketRequestID", ticketRequestController.CancelTicketRequest)
	r.PUT("/ticket-releases/:ticketReleaseID/ticket-requests/:ticketRequestID/add-ons", ticketRequestController.UpdateAddOns)

	// Ticket events routes
	r.GET("/events/:eventID/tickets", middleware.AuthorizeEventAccess(db, models.OrganizationMember), eventController.ListTickets)
	r.POST("/events/:eventID/tickets/qr-check-in", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketsController.QrCodeCheckIn)

	// My tickets
	r.GET("/my-ticket-requests", ticketRequestController.UsersList)
	r.GET("/my-tickets", ticketsController.UsersList)

	// Ticket routes
	r.DELETE("/my-tickets/:ticketID", ticketsController.CancelTicket)

	// send outs
	r.GET("/events/:eventID/send-outs", middleware.AuthorizeEventAccess(db, models.OrganizationMember), sendOutcontroller.GetEventSendOuts)
	r.POST("/events/:eventID/send-out", middleware.AuthorizeEventAccess(db, models.OrganizationMember), sendOutcontroller.SendOut)

	// Ticket routes
	r.GET("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketsController.GetTicket)
	r.PUT("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketsController.UpdateTicket)
	r.GET("/tickets/:ticketID/create-payment-intent", paymentsController.CreatePaymentIntent)

	// Sales report
	r.POST("/events/:eventID/sales-report", middleware.AuthorizeEventAccess(db, models.OrganizationMember), salesReportController.GenerateSalesReport)
	r.GET("/events/:eventID/sales-report", middleware.AuthorizeEventAccess(db, models.OrganizationMember), salesReportController.ListSalesReport)

	r.POST("/organizations", authentication.RequireRole(models.RoleSuperAdmin, db), organizationController.CreateOrganization)
	r.GET("/organizations", organizationController.ListOrganizations)
	r.GET("my-organizations", organizationController.ListMyOrganizations)
	r.GET("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db, models.OrganizationMember), organizationController.GetOrganization)
	r.PUT("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db, models.OrganizationMember), organizationController.UpdateOrganization)
	r.DELETE("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db, models.OrganizationMember), organizationController.DeleteOrganization)
	r.GET("/organizations/:organizationID/events", middleware.AuthorizeOrganizationAccess(db, models.OrganizationMember), organizationController.ListOrganizationEvents)

	// Organization Users routes
	r.GET("/organizations/:organizationID/users", middleware.AuthorizeOrganizationRole(db, models.OrganizationMember), organizationUsersController.GetOrganizationUsers)
	r.POST("/organizations/:organizationID/users/:username", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.AddUserToOrganization)
	r.DELETE("/organizations/:organizationID/users/:username", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.RemoveUserFromOrganization)
	r.PUT("/organizations/:organizationID/users/:username", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.ChangeUserOrganizationRole)

	// Ticket Release Methods routes
	r.GET("/ticket-release-methods", ticketReleaseMethodsController.ListTicketReleaseMethods)
	r.POST("/ticket-release-methods", authentication.RequireRole(models.RoleSuperAdmin, db), ticketReleaseMethodsController.CreateTicketReleaseMethod)

	// Ticket Types routes
	r.GET("/ticket-types",
		authentication.RequireRole(models.RoleSuperAdmin, db),
		ticketTypeController.ListAllTicketTypes)
	r.POST("/ticket-types", authentication.RequireRole(models.RoleSuperAdmin, db),
		ticketTypeController.CreateTicketTypes)

	// User Food Preference routes
	r.PUT("/user-food-preferences", userFoodPreferenceController.Update)
	r.GET("/user-food-preferences", userFoodPreferenceController.Get)
	r.GET("/food-preferences", userFoodPreferenceController.ListFoodPreferences)

	// Misc user
	r.PUT("/user/showed-post-login-screen", userController.UpdateShowedPostLogin)

	r.POST("/admin/create-user", authentication.RequireRole(models.RoleSuperAdmin, db), userController.CreateUser)

	// Banking details
	r.GET("/organizations/:organizationID/banking-details", middleware.AuthorizeOrganizationRole(db, models.OrganizationMember), bankingController.GetBankingDetails)
	r.POST("/organizations/:organizationID/banking-details", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), bankingController.SubmitBankingDetails)
	r.DELETE("/organizations/:organizationID/banking-details", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), bankingController.DeleteBankingDetails)

	r.POST("send-test-email", authentication.RequireRole(models.RoleSuperAdmin, db), notificationController.SendTestEmail)

	r = ManagerRoutes(r, db)
	r = PlanEnrollmentRoutes(r, db)
	r = AdminRoutes(r, db)

	return r
}
