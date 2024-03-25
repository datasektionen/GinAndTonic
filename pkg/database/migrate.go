package database

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/models/tr_methods"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Organization{},
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
		&models.OrganizationRole{},
		&models.OrganizationUserRole{},
		&models.Transaction{},
		&models.Notification{},
		&models.TicketReleaseReminder{},
		&models.UserPasswordReset{},
		&models.EventSalesReport{},
		&models.WebhookEvent{},
		&models.AddOn{},
		&models.TicketAddOn{},
		&tr_methods.LotteryConfig{},
	)
	return err
}
