package main

import (
	"github.com/gin-contrib/cors"
	_ "github.com/lib/pq"

	"log"

	"github.com/DowLucas/gin-ticket-release/pkg/database"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/routes"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file
	var err error

	if err = godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

}

func CORSConfig() cors.Config {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5000"}
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers", "Content-Type", "X-XSRF-TOKEN", "Accept", "Origin", "X-Requested-With", "Authorization")
	corsConfig.AddAllowMethods("GET", "POST", "PUT", "DELETE")
	return corsConfig
}

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
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

	router := routes.SetupRouter(db)

	router.Run()
}
