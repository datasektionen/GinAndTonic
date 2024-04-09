package models

import (
	"gorm.io/gorm"
)

type TRM string

const (
	FCFS_LOTTERY            TRM = "First Come First Serve Lottery"
	RESERVED_TICKET_RELEASE TRM = "Reserved Ticket Release"
	FCFS                    TRM = "First Come First Serve"
	SELECTIVE               TRM = "Selective"
)

type TicketReleaseMethod struct {
	gorm.Model
	MethodName  string `gorm:"unique" json:"method_name"`
	Description string `json:"description"`
}

func InitializeTicketReleaseMethods(db *gorm.DB) error {
	methods := []TicketReleaseMethod{
		{MethodName: string(FCFS_LOTTERY), Description: "First Come First Serve Lottery is a ticket release method where the people who requests a ticket within a specified time frame will be entered into a lottery. When tickets are allocated, all the ticket requests are entered into a lottery and the winners are selected randomly. The winners will be given a ticket and the rest will be put on the waitlist. Everyone who requested a ticket after the specified time frame will be put on the waitlist, unless the lottery isn't full. If the lottery isn't full, the remaining tickets will be given to the people on the waitlist, in the order they requested the ticket."},
		{MethodName: string(RESERVED_TICKET_RELEASE), Description: "Gives everyone in the ticket release a ticket"},
		{MethodName: string(FCFS), Description: "First Come First Serve will allocate tickets to the first people who requests a ticket. If the ticket release is full, the rest will be put on the waitlist."},
		{MethodName: string(SELECTIVE), Description: "Selective will allocate tickets to the people the organizer selects"},
	}

	for _, method := range methods {
		var existingMethod TicketReleaseMethod
		if err := db.Where("method_name = ?", method.MethodName).First(&existingMethod).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&method).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}
