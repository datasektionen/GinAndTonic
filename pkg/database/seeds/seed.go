package seeds

import (
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/database/factory"
	"gorm.io/gorm"
)

func SeedEvents(db *gorm.DB) {
	for i := 0; i < 10; i++ {
		event := factory.NewEvent(
			"Event Name",
			"Event Description",
			"Event Location",
			"CreatedByUser",
			1, // Example OrganizationID
			time.Now(),
		)
		db.Create(event)
	}
}

func SeedOrganizationRoles(db *gorm.DB) {
	roles := []string{"member", "owner"}
	for _, roleName := range roles {
		role := factory.NewOrganizationRole(roleName)
		db.Create(role)
	}
}

func SeedOrganizationUserRoles(db *gorm.DB) {
	userRole := factory.NewOrganizationUserRole(
		"UserUGKthID",
		1, // Example OrganizationID
		"member",
	)
	db.Create(userRole)
}

func SeedOrganizations(db *gorm.DB) {
	for i := 0; i < 5; i++ {
		org := factory.NewOrganization("Organization Name", "Organization Email")
		db.Create(org)
	}
}

func SeedTicketReleaseMethodDetails(db *gorm.DB) {
	detail := factory.NewTicketReleaseMethodDetail(
		10,
		"Email",
		"Standard",
		3600, // OpenWindowDuration in seconds
		1,    // Example TicketReleaseMethodID
	)
	db.Create(detail)
}

func SeedTicketReleaseMethods(db *gorm.DB) {
	method := factory.NewTicketReleaseMethod(
		"FCFS_LOTTERY",
		"First Come First Serve Lottery",
	)
	db.Create(method)
}

func SeedTicketReleases(db *gorm.DB) {
	release := factory.NewTicketRelease(
		1,          // Example EventID
		time.Now(), // Open
		time.Now(), // Close
		false,
		1, // Example TicketReleaseMethodDetailID
	)
	db.Create(release)
}

func SeedTicketRequests(db *gorm.DB) {
	request := factory.NewTicketRequest(
		2,
		1, // Example TicketReleaseID
		1, // Example TicketTypeID
		"UserUGKthID",
		false,
		time.Now(),
	)
	db.Create(request)
}

func SeedTicketTypes(db *gorm.DB) {
	ticketType := factory.NewTicketType(
		1, // Example EventID
		"General Admission",
		"Standard Ticket",
		50.00, // Price
		100,   // QuantityTotal
		false,
		1, // Example TicketReleaseID
	)
	db.Create(ticketType)
}

func SeedTickets(db *gorm.DB) {
	ticket := factory.NewTicket(
		1, // Example TicketRequestID
		false,
		false,
		"UserUGKthID",
	)
	db.Create(ticket)
}

func SeedUserFoodPreferences(db *gorm.DB) {
	var params factory.UserFoodPreferenceParams = factory.UserFoodPreferenceParams{
		UserUGKthID:       "UserUGKthID",
		Vegetarian:        false,
		Vegan:             false,
		GlutenIntolerant:  false,
		LactoseIntolerant: false,
		NutAllergy:        false,
		ShellfishAllergy:  false,
		PreferMeat:        false,
		AdditionalInfo:    "Additional Info",
	}

	preference := factory.NewUserFoodPreference(params)
	db.Create(preference)
}

func SeedUsers(db *gorm.DB) {
	user := factory.NewUser(
		"UGKthID",
		"Username",
		"FirstName",
		"LastName",
		"Email@example.com",
		1, // Example RoleID
	)
	db.Create(user)
}
