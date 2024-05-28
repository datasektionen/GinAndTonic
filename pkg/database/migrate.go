package database

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/models/tr_methods"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	fmt.Println("Migrating database...")

	if err := db.AutoMigrate(&models.User{}); err != nil {
		fmt.Println("error", err)
		return err
	}

	if err := db.AutoMigrate(
		&models.Network{},
		&models.NetworkDetails{},
		&models.NetworkMerchant{},
		&models.NetworkMerchantTerminals{},
	); err != nil {
		fmt.Println("error", err)
		return err
	}

	err := db.AutoMigrate(
		&models.PlanEnrollment{},
		&models.Organization{},
		&models.FeatureGroup{},
		&models.FeatureUsage{},
		&models.FeatureLimit{},
		&models.Feature{},
		&models.PackageTier{},
		&models.TicketReleaseMethod{},
		&models.TicketReleaseMethodDetail{},
		&models.TicketRelease{},
		&models.UserFoodPreference{},
		&models.EventFormField{},
		&models.NetworkUserRole{},
		&models.NetworkRole{},
		&models.EventFormFieldResponse{},
		&models.Event{},
		&models.EventLandingPage{},
		&models.TicketType{},
		&models.Ticket{},
		&models.TicketRequest{},
		&models.Role{},
		&models.OrganizationRole{},
		&models.OrganizationUserRole{},
		&models.Transaction{},
		&models.SendOut{},
		&models.Notification{},
		&models.TicketReleaseReminder{},
		&models.UserPasswordReset{},
		&models.EventSalesReport{},
		&models.WebhookEvent{},
		&models.AddOn{},
		&models.TicketAddOn{},
		&models.TicketReleasePaymentDeadline{},
		&models.EventSiteVisit{},
		&models.EventSiteVisitSummary{},
		&models.BankingDetail{},
		&models.ReferralSource{},
		&tr_methods.LotteryConfig{},
	)

	if err != nil {
		fmt.Println("error", err)
		return err
	}

	fmt.Println("Database migrated successfully")

	return nil
}
