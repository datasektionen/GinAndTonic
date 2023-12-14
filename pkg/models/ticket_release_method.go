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
		{MethodName: string(FCFS_LOTTERY), Description: "First Come First Serve Lottery"},
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
