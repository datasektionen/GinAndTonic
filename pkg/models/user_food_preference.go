package models

import "gorm.io/gorm"

type UserFoodPreference struct {
	gorm.Model
	UserUGKthID string `gorm:"primaryKey" json:"user_ug_kth_id"`
	User        User

	GlutenIntolerant  bool `json:"gluten_intolerant" gorm:"default:false"`
	LactoseIntolerant bool `json:"lactose_intolerant" gorm:"default:false"`
	Vegetarian        bool `json:"vegetarian" gorm:"default:false"`
	Vegan             bool `json:"vegan" gorm:"default:false"`
	NutAllergy        bool `json:"nut_allergy" gorm:"default:false"`
	ShellfishAllergy  bool `json:"shellfish_allergy" gorm:"default:false"`

	AdditionalInfo string `json:"additional_info" gorm:"default:''"`
}
