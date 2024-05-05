package database

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/models/tr_methods"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.PlanEnrollment{},
		&models.Feature{},
		&models.Network{},
		&models.NetworkUserRole{},
		&models.NetworkRole{},
		&models.PackageTier{},
		&models.FeatureUsage{},
		&models.FeatureLimit{},
		&models.Organization{},
		&models.TicketReleaseMethod{},
		&models.TicketReleaseMethodDetail{},
		&models.TicketRelease{},
		&models.UserFoodPreference{},
		&models.EventFormField{},
		&models.User{},
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
