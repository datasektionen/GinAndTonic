package models

import (
	"gorm.io/gorm"
)

type TRM string

const (
	FCFS_LOTTERY            TRM = "First Come First Serve Lottery"
	RESERVED_TICKET_RELEASE TRM = "Reserved Ticket Release"
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

// func (m *TicketReleaseMethod) SerializeConfig(config interface{}) error {
// 	data, err := json.Marshal(config)
// 	if err != nil {
// 		return err
// 	}
// 	m.ConfigData = string(data)
// 	return nil
// }

// func (m *TicketReleaseMethod) DeserializeConfig() (interface{}, error) {
// 	var config interface{}
// 	switch m.MethodType {
// 	case "Lottery":
// 		config = &tr_methods.LotteryConfig{}
// 	default:
// 		return nil, errors.New("unknown method type")
// 	}
// 	err := json.Unmarshal([]byte(m.ConfigData), &config)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return config, nil
// }
