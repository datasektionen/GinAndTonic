package controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConstantOptionsController struct {
}

// NewConstantOptionscontroller creates a new controller with the given database client
func NewConstantOptionsController(db *gorm.DB) *ConstantOptionsController {
	return &ConstantOptionsController{}
}

type Constants struct {
	CancellationPolicies []string `json:"cancellation_policies"`
	NotificationMethods  []string `json:"notification_methods"`
}

func (co *ConstantOptionsController) ListTicketReleaseConstants(c *gin.Context) {
	// may look confusing, but these are just constants defined in the models package
	constants := Constants{
		CancellationPolicies: []string{models.FULL_REFUND, models.NO_REFUND},
		NotificationMethods:  []string{models.EMAIL},
	}

	c.JSON(http.StatusOK, constants)
}
