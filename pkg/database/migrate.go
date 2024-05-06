package database

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/models/tr_methods"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Network{},
		&models.PlanEnrollment{},
		&models.User{},
		&models.FeatureGroup{},
		&models.FeatureUsage{},
		&models.FeatureLimit{},
		&models.Feature{},
		&models.PackageTier{},
		&models.TicketReleaseMethod{},
		&models.Organization{},
		&models.TicketReleaseMethodDetail{},
		&models.TicketRelease{},
		&models.UserFoodPreference{},
		&models.EventFormField{},
		&models.NetworkUserRole{},
		&models.NetworkRole{},
		&models.EventFormFieldResponse{},
		&models.Event{},
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
	return err
}
