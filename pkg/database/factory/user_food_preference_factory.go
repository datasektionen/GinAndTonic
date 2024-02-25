package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

type UserFoodPreferenceParams struct {
	UserUGKthID       string
	GlutenIntolerant  bool
	LactoseIntolerant bool
	Vegetarian        bool
	Vegan             bool
	NutAllergy        bool
	ShellfishAllergy  bool
	AdditionalInfo    string
	PreferMeat        bool
}

func NewUserFoodPreference(params UserFoodPreferenceParams) *models.UserFoodPreference {
	return &models.UserFoodPreference{
		UserUGKthID:       params.UserUGKthID,
		GlutenIntolerant:  params.GlutenIntolerant,
		LactoseIntolerant: params.LactoseIntolerant,
		Vegetarian:        params.Vegetarian,
		Vegan:             params.Vegan,
		NutAllergy:        params.NutAllergy,
		ShellfishAllergy:  params.ShellfishAllergy,
		AdditionalInfo:    params.AdditionalInfo,
		PreferMeat:        params.PreferMeat,
	}
}
