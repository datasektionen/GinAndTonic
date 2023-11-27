package testutils

import (
	"fmt"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/database/factory"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/models/tr_methods"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var modelslist = []interface{}{
	&models.Organization{},
	&models.TicketReleaseMethod{},
	&models.TicketReleaseMethodDetail{},
	&models.TicketRelease{},
	&models.User{},
	&models.UserFoodPreference{},
	&models.Event{},
	&models.TicketType{},
	&models.Ticket{},
	&models.TicketRequest{},
	&models.Role{},
	&models.OrganizationRole{},
	&models.OrganizationUserRole{},
	&tr_methods.LotteryConfig{},
}

func SetupTestDatabase() (*gorm.DB, error) {
	// Connect to an in-memory SQLite database

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.Exec("PRAGMA foreign_keys = ON")

	// Run migrations
	err = db.AutoMigrate(modelslist...)

	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	err = models.InitializeRoles(db)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize roles: %w", err)
	}

	err = models.InitializeOrganizationRoles(db)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize organization roles: %w", err)
	}

	return db, nil
}

func CleanupTestDatabase(db *gorm.DB) {
	// Drop all tables
	db.Migrator().DropTable(modelslist...)
}

func CreateTicketRelease(totalTickets int, methodName models.TRM, openWindowDuration int64, openTime int64) models.TicketRelease {
	return models.TicketRelease{
		TicketTypes: []models.TicketType{
			{
				QuantityTotal: uint(totalTickets),
			},
		},
		TicketReleaseMethodDetail: models.TicketReleaseMethodDetail{
			TicketReleaseMethod: models.TicketReleaseMethod{
				MethodName: string(methodName),
			},
			OpenWindowDuration: uint(openWindowDuration),
		},
		Open: uint(openTime),
	}
}

func SetupUserWorkflow(db *gorm.DB) {
	// Seed the database with necessary data for testing
	// For example, create TicketRelease, TicketType, User, etc.
	// Create a Event
	user := models.User{
		UGKthID: "validUserUGKthID",
	}

	db.Create(&user)
}

func SetupOrganizationWorkflow(db *gorm.DB) {
	// Seed the database with necessary data for testing
	// For example, create TicketRelease, TicketType, User, etc.
	// Create a Event

	user := models.User{
		UGKthID:   "validUserUGKthID",
		Username:  "validUsername",
		Email:     "validEmail",
		FirstName: "validFirstName",
		LastName:  "validLastName",
		RoleID:    1, // User role
	}

	db.Create(&user)

	organization := models.Organization{
		Name: "validOrganizationName",
	}

	db.Create(&organization)

	//Associate user with organization
	db.Model(&organization).Association("Users").Append(&user)
}

func SetupEventWorkflow(db *gorm.DB) {
	// Seed the database with necessary data for testing
	// For example, create TicketRelease, TicketType, User, etc.
	// Create a Event
	event := models.Event{
		Name:        "validEventName",
		Description: "validEventDescription",
		Location:    "validEventLocation",
		// Date is a time.Time type
		Date:           time.Now(),
		OrganizationID: 1,
		CreatedBy:      "validUserUGKthID",
	}

	db.Create(&event)

	// Create ticket release method
	ticketReleaseMethod := factory.NewTicketReleaseMethod(
		string(models.FCFS_LOTTERY),
		"validTicketReleaseMethodDescription",
	)

	db.Create(ticketReleaseMethod)

	ticketReleaseMethodDetail := factory.NewTicketReleaseMethodDetail(
		10,
		"Email",
		"Standard",
		3600, // OpenWindowDuration in seconds
		1,    // Example TicketReleaseMethodID
	)

	db.Create(ticketReleaseMethodDetail)

	// Create ticket types

	// Create a TicketRelease
	ticketRelease := models.TicketRelease{
		EventID:                     int(event.ID),
		Open:                        uint(time.Now().Unix()) - 1000,
		Close:                       uint(time.Now().Unix()) + 1000,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetail.ID,
	}

	ticketType := factory.NewTicketType(event.ID, "validTicketTypeName", "validTicketTypeDescription", 100, 100, false, 1)

	db.Create(&ticketRelease)
	db.Create(ticketType)
}
