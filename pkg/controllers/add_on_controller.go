package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AddOnController struct {
	DB *gorm.DB
}

// NewAddOnController creates a new controller with the given database client
func NewAddOnController(db *gorm.DB) *AddOnController {
	return &AddOnController{DB: db}
}

func (aoc *AddOnController) GetAddOns(c *gin.Context) {
	ticketReleaseID, err := strconv.Atoi(c.Param("ticketReleaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var addOns []models.AddOn
	if err := aoc.DB.Where("ticket_release_id = ?", ticketReleaseID).Find(&addOns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"add_ons": addOns})
}

func (aoc *AddOnController) UpsertAddOns(c *gin.Context) {
	var input []models.AddOn
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketReleaseID, err := strconv.Atoi(c.Param("ticketReleaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := aoc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := aoc.deleteNonExistingAddOns(tx, ticketReleaseID, input); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := aoc.upsertAddOns(tx, ticketReleaseID, input); err != nil {
		tx.Rollback()

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "AddOns upserted successfully"})
}

func (aoc *AddOnController) deleteNonExistingAddOns(tx *gorm.DB, ticketReleaseID int, input []models.AddOn) error {
	var existingAddOns []models.AddOn
	tx.Where("ticket_release_id = ?", ticketReleaseID).Find(&existingAddOns)

	// Convert input to map for faster lookup
	inputMap := make(map[string]models.AddOn)
	for _, addOnInput := range input {
		inputMap[addOnInput.Name] = addOnInput
	}

	// Delete AddOns that no longer exist in the input
	for _, existingAddOn := range existingAddOns {
		if _, exists := inputMap[existingAddOn.Name]; !exists {
			tx.Unscoped().Delete(&existingAddOn)
		}
	}

	return nil
}

func (aoc *AddOnController) upsertAddOns(tx *gorm.DB, ticketReleaseID int, input []models.AddOn) error {
	for _, addOnInput := range input {
		var addOn models.AddOn
		if err := tx.Where("name = ? AND ticket_release_id = ?", addOnInput.Name, ticketReleaseID).First(&addOn).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Record not found, create a new one
				if err := aoc.createAddOn(tx, ticketReleaseID, addOnInput); err != nil {
					return err
				}
			} else {
				// Some other error occurred
				fmt.Println(err)
				return err
			}
		} else {
			// Record found, update it
			if err := aoc.updateAddOn(tx, ticketReleaseID, addOn, addOnInput); err != nil {
				return err
			}
		}
	}

	return nil
}

func (aoc *AddOnController) createAddOn(tx *gorm.DB, ticketReleaseID int, addOnInput models.AddOn) error {
	addOn := models.AddOn{
		Name:            addOnInput.Name,
		Description:     addOnInput.Description,
		Price:           addOnInput.Price,
		MaxQuantity:     addOnInput.MaxQuantity,
		IsEnabled:       addOnInput.IsEnabled,
		TicketReleaseID: ticketReleaseID,
	}

	if err := addOn.ValidateAddOn(); err != nil {
		return errors.New("invalid addOn, " + err.Error())
	}

	if err := tx.Create(&addOn).Error; err != nil {
		return err // This ensures the transaction is properly rolled back in the calling function
	}

	return nil
}

func (aoc *AddOnController) updateAddOn(tx *gorm.DB, ticketReleaseID int, addOn models.AddOn, addOnInput models.AddOn) error {
	updatedAddOn := models.AddOn{
		Name:            addOnInput.Name,
		Description:     addOnInput.Description,
		Price:           addOnInput.Price,
		MaxQuantity:     addOnInput.MaxQuantity,
		IsEnabled:       addOnInput.IsEnabled,
		TicketReleaseID: ticketReleaseID,
	}

	if err := updatedAddOn.ValidateAddOn(); err != nil {
		return errors.New("invalid addOn, " + err.Error())
	}

	if err := tx.Model(&addOn).Updates(updatedAddOn).Error; err != nil {
		fmt.Println("Error updating addOn:", err)
		return err // Handle the error appropriately
	}

	return nil
}
