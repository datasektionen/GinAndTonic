package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/DowLucas/gin-ticket-release/pkg/database"
	"github.com/DowLucas/gin-ticket-release/pkg/jobs"
	"github.com/DowLucas/gin-ticket-release/pkg/jobs/tasks"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/routes"
)

var log = logrus.New()

func createLogDirAndLogFiles() {
	// Create logs directory if it doesn't exist
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	// Create log files if they don't exist
	if _, err := os.Stat("logs/allocate_reserve_tickets_job.log"); os.IsNotExist(err) {
		os.Create("logs/allocate_reserve_tickets_job.log")
	}

	if _, err := os.Stat("logs/notification.log"); os.IsNotExist(err) {
		os.Create("logs/notification.log")
	}
}

func init() {
	createLogDirAndLogFiles()

	// Load environment variables from .env file
	var err error
	if os.Getenv("ENV") == "dev" {
		if err = godotenv.Load(".env"); err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Error loading .env file")

		}

	}

	// Set log output to the file
	log.SetOutput(os.Stdout)

	// Set log level
	log.SetLevel(logrus.InfoLevel)

	// Log as JSON for structured logging
	log.SetFormatter(&logrus.JSONFormatter{})
}

func CORSConfig() cors.Config {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5000", "http://localhost", "http://localhost:8080", "http://tessera.datasektionen.se"}
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers", "Content-Type", "X-XSRF-TOKEN", "Accept", "Origin", "X-Requested-With", "Authorization")
	corsConfig.AddAllowMethods("GET", "POST", "PUT", "DELETE")
	return corsConfig
}

func setupCronJobs(db *gorm.DB) *cron.Cron {
	c := cron.New()
	_, err := c.AddFunc("@every 30m", func() {
		jobs.AllocateReserveTicketsJob(db)
	})

	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to add AllocateReserveTicketsJob to cron")
	}

	_, err = c.AddFunc("@every 24h", func() {
		jobs.NotifyReserveNumberJob(db)
	})

	// Run 1 month before the end of year
	_, err = c.AddFunc("0 0 1 12 *", func() {
		jobs.GDPRRenewalNotifyJob(db)
	})

	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to add GDPRRenewalNotifyJob to cron")
	}
	
	_, err = c.AddFunc("0 0 1 1 *", func() {
		jobs.GDPRCheckRenewalJob(db)
	})

	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to add GDPRCheckRenewalJob to cron")
	}

	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to add NotifyReserveNumberJob to cron")
	}

	fmt.Println("Starting cron jobs")
	c.Start()

	return c
}

func startAsynqServer(db *gorm.DB) *asynq.Server {
	// Parse the REDIS_URL
	redisURL, err := url.Parse(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("Failed to parse REDIS_URL: %v", err)
	}

	// Extract host and password. Port is part of the host in the URL.
	redisHost := redisURL.Host
	redisPassword, _ := redisURL.User.Password()

	// Create a new Asynq server instance with RedisClientOpt.
	var srv *asynq.Server

	fmt.Println("Starting Asynq server...")
	if os.Getenv("ENV") == "dev" {
		srv = asynq.NewServer(
			asynq.RedisClientOpt{
				Addr: os.Getenv("REDIS_URL"),
			},
			asynq.Config{
				Concurrency: 1,
				Queues: map[string]int{
					"default": 1,
				},
			},
		)
	} else {
		srv = asynq.NewServer(
			asynq.RedisClientOpt{
				Addr:     redisHost,
				Password: redisPassword, // Set the password here.
			},
			asynq.Config{
				Concurrency: 5,
				Queues: map[string]int{
					"critical": 6,
					"default":  3,
					"low":      1,
				},
			},
		)
	}

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeEmail, jobs.HandleEmailJob(db))
	mux.HandleFunc(tasks.TypeReminderEmail, jobs.HandleReminderJob(db))

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatalf("Could not run Asynq server: %v", err)
			// For debugging, you might want to print the error and stop the execution:
			os.Exit(1)
		}
	}()

	return srv
}

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to connect to database")
	}

	createLogDirAndLogFiles()

	err = models.CreateOrganizationUniqueIndex(db)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to create unique index for organizations")
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Initialize roles

	if err := models.InitializeRoles(db); err != nil {
		panic("Failed to initialize roles: " + err.Error())
	}

	if err := models.InitializeOrganizationRoles(db); err != nil {
		panic("Failed to initialize organization roles: " + err.Error())
	}

	if err := models.InitializeTicketReleaseMethods(db); err != nil {
		panic("Failed to initialize ticket release methods: " + err.Error())
	}

	gin.SetMode(gin.ReleaseMode)

	// Setup cron jobs
	c := setupCronJobs(db)

	// jobs.StartEmailJobs(db, 5)
	asynqServer := startAsynqServer(db)

	router := routes.SetupRouter(db)

	// Create a server using router
	srv := &http.Server{
		Addr:    ":8080", // or your specific port
		Handler: router,
	}

	// Run the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to run Gin server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down servers...")

	// Shutdown cron jobs
	c.Stop()

	// Shutdown Asynq server
	asynqServer.Shutdown()

	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Server forced to shutdown")
	}

	log.Info("Servers gracefully stopped")
}
