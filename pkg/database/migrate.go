package database

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/models/tr_methods"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Team{},
		&models.TicketReleaseMethod{},
		&models.TicketReleaseMethodDetail{},
		&models.TicketRelease{},
		&models.PreferredEmail{},
		&models.User{},
		&models.UserFoodPreference{},
		&models.EventFormField{},
		&models.EventFormFieldResponse{},
		&models.Event{},
		&models.TicketType{},
		&models.Ticket{},
		&models.TicketRequest{},
		&models.Role{},
		&models.TeamRole{},
		&models.TeamUserRole{},
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
		&tr_methods.LotteryConfig{},
	)
	return err
}
