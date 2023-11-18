package services

import (
	"log"
	"os"
	"testing"

	"github.com/DowLucas/gin-ticket-release/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	setupDatabase()
	setupOrganizationService()
	setupTicketRequestService()
	// Run all the tests.
	code := m.Run()
	// Teardown code here (if necessary).
	os.Exit(code)
}

func setupDatabase() {
	var err error
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}

func setupOrganizationService() {
	SetupOrganizationDB()
}

func setupTicketRequestService() {
	SetupTicketRequestDB()
}
