package controllers

import (
	"fmt"
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserFoodPreferenceController struct {
	DB *gorm.DB
}

func NewUserFoodPreferenceController(db *gorm.DB) *UserFoodPreferenceController {
	return &UserFoodPreferenceController{DB: db}
}

func (ctrl *UserFoodPreferenceController) Update(c *gin.Context) {
	var userFoodPreference models.UserFoodPreference

	UGKthID, _ := c.Get("ugkthid")

	if err := c.ShouldBindJSON(&userFoodPreference); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the existing record
	var existingUserFoodPreference models.UserFoodPreference
	if err := ctrl.DB.Where("user_ug_kth_id = ?", UGKthID).First(&existingUserFoodPreference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error getting the user food preference"})
		return
	}

	existingUserFoodPreference.GlutenIntolerant = userFoodPreference.GlutenIntolerant
	existingUserFoodPreference.LactoseIntolerant = userFoodPreference.LactoseIntolerant
	existingUserFoodPreference.Vegetarian = userFoodPreference.Vegetarian
	existingUserFoodPreference.Vegan = userFoodPreference.Vegan
	existingUserFoodPreference.NutAllergy = userFoodPreference.NutAllergy
	existingUserFoodPreference.ShellfishAllergy = userFoodPreference.ShellfishAllergy
	existingUserFoodPreference.PreferMeat = userFoodPreference.PreferMeat
	existingUserFoodPreference.Kosher = userFoodPreference.Kosher
	existingUserFoodPreference.Halal = userFoodPreference.Halal
	existingUserFoodPreference.AdditionalInfo = userFoodPreference.AdditionalInfo
	existingUserFoodPreference.GDPRAgreed = userFoodPreference.GDPRAgreed

	if userFoodPreference.NeedsToRenewGDPR {
		existingUserFoodPreference.NeedsToRenewGDPR = false
	}

	if err := ctrl.DB.Save(&existingUserFoodPreference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error saving the user food preference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_food_preference": existingUserFoodPreference})
}

func (ctrl *UserFoodPreferenceController) Get(c *gin.Context) {
	var userFoodPreference models.UserFoodPreference

	UGKthID, _ := c.Get("ugkthid")

	if err := ctrl.DB.Where("user_ug_kth_id = ?", UGKthID).First(&userFoodPreference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error getting the user food preference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_food_preference": userFoodPreference})
}

func (ctrl *UserFoodPreferenceController) GuestGet(c *gin.Context) {
	fmt.Println("GuestGet")
	requestToken := c.Query("request_token")
	if requestToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing request token"})
		return
	}

	ugKTHId := c.Param("ugkthid")
	if ugKTHId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	var user models.User
	if err := ctrl.DB.Where("ug_kth_id = ? AND request_token = ?", ugKTHId, requestToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid request token"})
		return
	}

	if !user.IsGuestCustomer(ctrl.DB) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a guest"})
		return
	}

	var userFoodPreference models.UserFoodPreference

	if err := ctrl.DB.Where("user_ug_kth_id = ?", user.UGKthID).First(&userFoodPreference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error getting the user food preference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_food_preference": userFoodPreference})
}

func (ctrl *UserFoodPreferenceController) GuestUpdate(c *gin.Context) {
	var userFoodPreference models.UserFoodPreference

	if err := c.ShouldBindJSON(&userFoodPreference); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ugKTHId := c.Param("ugkthid")
	if ugKTHId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	requestToken := c.Query("request_token")
	if requestToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing request token"})
		return
	}

	var user models.User
	if err := ctrl.DB.Where("ug_kth_id = ? AND request_token = ?", ugKTHId, requestToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid request token"})
		return
	}

	if !user.IsGuestCustomer(ctrl.DB) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a guest"})
		return
	}

	// Find the existing record
	var existingUserFoodPreference models.UserFoodPreference
	if err := ctrl.DB.Where("user_ug_kth_id = ?", user.UGKthID).First(&existingUserFoodPreference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error getting the user food preference"})
		return
	}

	existingUserFoodPreference.GlutenIntolerant = userFoodPreference.GlutenIntolerant
	existingUserFoodPreference.LactoseIntolerant = userFoodPreference.LactoseIntolerant
	existingUserFoodPreference.Vegetarian = userFoodPreference.Vegetarian
	existingUserFoodPreference.Vegan = userFoodPreference.Vegan
	existingUserFoodPreference.NutAllergy = userFoodPreference.NutAllergy
	existingUserFoodPreference.ShellfishAllergy = userFoodPreference.ShellfishAllergy
	existingUserFoodPreference.PreferMeat = userFoodPreference.PreferMeat
	existingUserFoodPreference.Kosher = userFoodPreference.Kosher
	existingUserFoodPreference.Halal = userFoodPreference.Halal
	existingUserFoodPreference.AdditionalInfo = userFoodPreference.AdditionalInfo
	existingUserFoodPreference.GDPRAgreed = userFoodPreference.GDPRAgreed

	if userFoodPreference.NeedsToRenewGDPR {
		existingUserFoodPreference.NeedsToRenewGDPR = false
	}

	if err := ctrl.DB.Save(&existingUserFoodPreference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error saving the user food preference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_food_preference": existingUserFoodPreference})
}

func (ctrl *UserFoodPreferenceController) ListFoodPreferences(c *gin.Context) {
	// Use this to get all food preferences from the database
	// we are only interested in the name of the food preference
	alternatives := models.GetFoodPreferencesAlternatives()

	c.JSON(http.StatusOK, gin.H{"food_preferences": alternatives})
}

type UserFoodPreferenceRenewal struct {
	Renew bool `json:"renew"`
}
