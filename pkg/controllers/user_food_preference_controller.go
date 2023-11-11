package controllers

import (
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

	// Update user food preference
	if err := ctrl.DB.Model(&models.UserFoodPreference{}).Where("user_ug_kth_id = ?", UGKthID).Updates(&userFoodPreference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error updating the user food preference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_food_preference": userFoodPreference})
}

func (ctrl *UserFoodPreferenceController) Get(c *gin.Context) {
	var userFoodPreference models.UserFoodPreference

	UGKthID, _ := c.Get("ugkthid")

	if err := ctrl.DB.Preload("User").Where("user_ug_kth_id = ?", UGKthID).First(&userFoodPreference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error getting the user food preference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_food_preference": userFoodPreference})
}
