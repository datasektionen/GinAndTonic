package models

type UserUnlockedTicketRelease struct {
	UserUGKthID     string `gorm:"primaryKey"`
	TicketReleaseID uint   `gorm:"primaryKey"`
}
