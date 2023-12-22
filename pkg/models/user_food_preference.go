package models

import "gorm.io/gorm"

type UserFoodPreference struct {
	gorm.Model
	UserUGKthID       string `gorm:"primaryKey" json:"user_ug_kth_id"`
	GlutenIntolerant  bool   `json:"gluten_intolerant" gorm:"default:false"`
	LactoseIntolerant bool   `json:"lactose_intolerant" gorm:"default:false"`
	Vegetarian        bool   `json:"vegetarian" gorm:"default:false"`
	Vegan             bool   `json:"vegan" gorm:"default:false"`
	NutAllergy        bool   `json:"nut_allergy" gorm:"default:false"`
	ShellfishAllergy  bool   `json:"shellfish_allergy" gorm:"default:false"`
	Halal             bool   `json:"halal" gorm:"default:false"`
	Kosher            bool   `json:"kosher" gorm:"default:false"`
	AdditionalInfo    string `json:"additional_info" gorm:"default:''"`
}

// Function retrieving the Food preferences field of the UserFoodPreference struct
func GetFoodPreferencesAlternatives() []string {
	return []string{
		"gluten_intolerant",
		"lactose_intolerant",
		"vegetarian",
		"vegan",
		"nut_allergy",
		"shellfish_allergy",
		"halal",
		"kosher",
		"additional_info",
	}
}
