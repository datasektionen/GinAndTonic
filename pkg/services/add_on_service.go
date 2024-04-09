package services

import (
	"fmt"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"gorm.io/gorm"
)

func ValidateAddOnsForTicketRequest(db *gorm.DB, selectedAddOns []types.SelectedAddOns, ticketReleaseID int) *types.ErrorResponse {
	// Fetch all add-ons associated with the ticket release.
	var ticketReleaseAddOns []models.AddOn
	if err := db.Where("ticket_release_id = ?", ticketReleaseID).Find(&ticketReleaseAddOns).Error; err != nil {
		return &types.ErrorResponse{StatusCode: 500, Message: "Error getting ticket release add-ons"}
	}

	// Map add-ons by ID for easier lookup.
	addOnsMap := make(map[int]models.AddOn)
	for _, addOn := range ticketReleaseAddOns {
		addOnsMap[int(addOn.ID)] = addOn
	}

	// Validate each selected add-on.
	for _, selectedAddOn := range selectedAddOns {
		addOn, exists := addOnsMap[selectedAddOn.ID]
		if !exists {
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Add-on is not available for this ticket release")}
		}

		// Check if the add-on is enabled.
		if !addOn.IsEnabled {
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Add-on is not enabled")}
		}

		// Validate requested quantity against MaxQuantity.
		if selectedAddOn.Quantity > addOn.MaxQuantity {
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Requested quantity for add-on exceeds maximum allowed")}
		}

		if selectedAddOn.Quantity <= 0 {
			return &types.ErrorResponse{StatusCode: 400, Message: fmt.Sprintf("Requested quantity for add-on  must be greater than 0")}
		}
	}

	// If all validations pass, return nil.
	return nil
}
