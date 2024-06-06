package routes

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/controllers"
	surfboard_controllers "github.com/DowLucas/gin-ticket-release/pkg/controllers/surfboard"
	"github.com/DowLucas/gin-ticket-release/pkg/middleware"
	surfboard_middlware "github.com/DowLucas/gin-ticket-release/pkg/middleware/surfboard"
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
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept", "Authorization", "Range", "X-Requested-With"},
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

	r.POST("/signup", customerAuthController.SignupCustomerUser)
	r.POST("/login", customerAuthController.LoginUser)
	r.POST("/verify-email", customerAuthController.VerifyEmail)
	r.POST("/resend-verification-email", customerAuthController.ResendVerificationEmail)

	// Password reset
	r.POST("/password-reset", passwordResetController.CreatePasswordReset)
	r.POST("/password-reset/complete", passwordResetController.CompletePasswordReset)

	// Deprecated routes, to be removed
	// r.GET("/login", controllers.Login)
	// r.GET("/login-complete/:token", controllers.LoginComplete)

	r.GET("/current-user", authentication.ValidateTokenMiddleware(true), controllers.CurrentUser)
	r.GET("/logout", authentication.ValidateTokenMiddleware(true), controllers.Logout)

	eventController := controllers.NewEventController(db)

	userController := controllers.NewUserController(db)
	allocateTicketsService := services.NewAllocateTicketsService(db)
	bankingService := banking_service.NewBankingService(db)

	organizationController := controllers.NewOrganizationController(db)
	ticketReleaseMethodsController := controllers.NewTicketReleaseMethodsController(db)
	ticketReleaseController := controllers.NewTicketReleaseController(db)
	ticketReleasePromoCodeController := controllers.NewTicketReleasePromoCodeController(db)
	ticketTypeController := controllers.NewTicketTypeController(db)
	organizationUsersController := controllers.NewOrganizationUsersController(db)
	userFoodPreferenceController := controllers.NewUserFoodPreferenceController(db)
	ticketOrderController := controllers.NewTicketOrderController(db)
	allocateTicketsController := controllers.NewAllocateTicketsController(db, allocateTicketsService)
	ticketsController := controllers.NewTicketController(db)
	constantOptionsController := controllers.NewConstantOptionsController(db)
	notificationController := controllers.NewNotificationController(db)
	contactController := controllers.NewContactController(db)
	ticketReleaseReminderController := controllers.NewTicketReleaseReminderController(db)
	eventFormFieldController := controllers.NewEventFormFieldController(db)
	eventFromFieldResponseController := controllers.NewEventFormFieldResponseController(db)
	addOnController := controllers.NewAddOnController(db)
	eventSiteVistsController := controllers.NewSitVisitsController(db)
	bankingController := controllers.NewBankingController(bankingService)
	guestController := controllers.NewGuestController(db)
	surfboardPaymentController := surfboard_controllers.NewPaymentSurfboardController(db)
	surfboardPaymentWebhookController := surfboard_controllers.NewPaymentWebhookController(db)

	var rlm *RateLimiterMiddleware
	var rlmURLParam *RateLimiterMiddleware
	if os.Getenv("ENV") == "dev" {
		// For development, we dont really care about the rate limit
		// 1 request per second
		rlm = NewRateLimiterMiddleware(rate.Limit(1), 1)
		rlmURLParam = NewRateLimiterMiddleware(2, 5)
	} else {
		rlm = NewRateLimiterMiddleware(rate.Limit(1.0/60.0), 1)
		rlmURLParam = NewRateLimiterMiddleware(rate.Limit(1.0/60.0), 1)
	}

	r.GET("/ticket-release/constants", constantOptionsController.ListTicketReleaseConstants)
	r.POST("/payments/webhook", surfboard_middlware.ValidatePaymentWebhookSignature(), surfboardPaymentWebhookController.HandlePaymentWebhook)

	r.GET("/view/events/:refID", authentication.ValidateTokenMiddleware(false), middleware.UpdateSiteVisits(db), eventController.GetEvent)
	r.GET("/view/events/:refID/landing-page", eventController.GetUsersView)
	r.GET("/timestamp", eventController.GetTimestamp)
	r.GET("/guest-customer/:ugkthid/activate-promo-code/:eventID", ticketReleasePromoCodeController.GuestCreate)
	r.GET("/guest-customer/:ugkthid", guestController.Get)
	r.DELETE("/guest-customer/:ugkthid/ticket-requests/:ticketOrderID", ticketOrderController.GuestCancelTicketOrder)
	r.DELETE("/guest-customer/:ugkthid/my-tickets/:ticketID", ticketsController.GuestCancelTicket)
	r.PUT("/guest-customer/:ugkthid/events/:eventID/ticket-requests/:ticketOrderID/form-fields", eventFromFieldResponseController.GuestUpsert)
	r.GET("/guest-customer/:ugkthid/user-food-preferences", userFoodPreferenceController.GuestGet)
	r.PUT("/guest-customer/:ugkthid/user-food-preferences", userFoodPreferenceController.GuestUpdate)
	r.POST("/guest-customer/:ugkthid/events/:eventID/guest-customer/ticket-requests", rlmURLParam.MiddlewareFuncURLParam(), ticketOrderController.GuestCreate)

	r.Use(authentication.ValidateTokenMiddleware(true))
	r.Use(middleware.UserLoader(db))

	synqMonHandler := setupAsynqMon()
	r.Any("/admin/monitoring/*any", authentication.RequireRole(models.RoleSuperAdmin, db), gin.WrapH(synqMonHandler)) // Serve asynqmon on /monitoring path

	//Event routes

	r.GET("/events", authentication.RequireRole(models.RoleSuperAdmin, db), eventController.ListEvents)
	r.GET("/events/:eventID", middleware.UpdateSiteVisits(db), eventController.GetEvent)
	r.GET("/events/:eventID/manage",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember), gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "User has access to this event"})
		}))

	r.GET("/test", authentication.RequireRole(models.RoleSuperAdmin, db), func(c *gin.Context) {
		var network models.Network
		if err := db.Preload("Organizations.Store").Preload("Merchant").Preload("Details").First(&network).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Terminal created"})
	})

	r.GET("/events/:eventID/manage/secret-token",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventController.GetEventSecretToken)

	r.PUT("/events/:eventID",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		eventController.UpdateEvent)

	// Contact
	r.POST("/contact", contactController.CreateContact)
	r.POST("/plan-contact", contactController.CreatePlanContact)

	// Ticket release routes
	r.GET("/events/:eventID/ticket-release", ticketReleaseController.ListEventTicketReleases)
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

	r.PUT("/templates/ticket-release/:ticketReleaseID/unsave", ticketReleaseController.UnsaveTemplate)
	r.GET("/templates/ticket-release", ticketReleaseController.GetTemplateTicketReleases)

	r.PUT("/templates/ticket-types/:ticketTypeID/unsave", ticketTypeController.UnsaveTemplate)
	r.GET("/templates/ticket-types", ticketTypeController.GetTemplateTicketTypes)

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
	r.PUT("/events/:eventID/ticket-requests/:ticketOrderID/form-fields", eventFromFieldResponseController.Upsert)

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
	r.POST("/events/:eventID/ticket-requests/:ticketOrderID/allocate",
		middleware.AuthorizeEventAccess(db, models.OrganizationMember),
		allocateTicketsController.SelectivelyAllocateTicketOrder)

	// Ticket request event routes
	r.GET("/events/:eventID/ticket-requests", ticketOrderController.Get)
	r.POST("/events/:eventID/ticket-requests", rlm.MiddlewareFunc(), ticketOrderController.Create)
	r.DELETE("/events/:eventID/ticket-requests/:ticketOrderID", ticketOrderController.CancelTicketOrder)
	r.PUT("/ticket-releases/:ticketReleaseID/ticket-requests/:ticketOrderID/add-ons", ticketOrderController.UpdateAddOns)

	// Ticket events routes
	r.GET("/events/:eventID/tickets", middleware.AuthorizeEventAccess(db, models.OrganizationMember), eventController.ListTickets)

	// My tickets
	r.GET("/my-ticket-requests", ticketOrderController.UsersList)
	r.GET("/my-tickets", ticketsController.UsersList)

	// Ticket routes
	r.DELETE("/my-tickets/:ticketID", ticketsController.CancelTicket)

	// Ticket routes
	r.GET("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketsController.GetTicket)
	r.PUT("/events/:eventID/tickets/:ticketID", middleware.AuthorizeEventAccess(db, models.OrganizationMember), ticketsController.UpdateTicket)

	// All routes that have with payments

	// Route for creating an order for a list of tickets
	r.POST("/payments/events/:referenceID/order/create", surfboardPaymentController.CreateOrder)
	r.POST("/payments/events/:referenceID/order/:orderID/status", surfboardPaymentController.GetOrderStatus)

	r.GET("/organizations", organizationController.ListOrganizations)
	r.GET("my-organizations", organizationController.ListMyOrganizations)
	r.GET("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db, models.OrganizationMember), organizationController.GetOrganization)
	r.PUT("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db, models.OrganizationMember), organizationController.UpdateOrganization)
	r.DELETE("/organizations/:organizationID", middleware.AuthorizeOrganizationAccess(db, models.OrganizationMember), organizationController.DeleteOrganization)
	r.GET("/organizations/:organizationID/events", middleware.AuthorizeOrganizationAccess(db, models.OrganizationMember), organizationController.ListOrganizationEvents)

	// Organization Users routes
	r.GET("/organizations/:organizationID/users", middleware.AuthorizeOrganizationRole(db, models.OrganizationMember), organizationUsersController.GetOrganizationUsers)
	r.POST("/organizations/:organizationID/users", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.AddUserToOrganization)
	r.DELETE("/organizations/:organizationID/users", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.RemoveUserFromOrganization)
	r.PUT("/organizations/:organizationID/users", middleware.AuthorizeOrganizationRole(db, models.OrganizationOwner), organizationUsersController.ChangeUserOrganizationRole)

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
