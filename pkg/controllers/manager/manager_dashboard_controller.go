package manager_controller

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	manager_service "github.com/DowLucas/gin-ticket-release/pkg/services/manager"
	response_utils "github.com/DowLucas/gin-ticket-release/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
ManagerController is a struct that holds the methods for the manager controller.
*/

type ManagerController struct {
	DB              *gorm.DB
	service         *manager_service.ManagerService
	network_service *manager_service.NetworkService
}

func NewManagerController(db *gorm.DB) *ManagerController {
	return &ManagerController{
		DB:              db,
		service:         manager_service.NewManagerService(db),
		network_service: manager_service.NewNetworkService(db),
	}
}

func (mc *ManagerController) GetNetworkDetails(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	network, rerr := mc.network_service.GetNetworkDetails(&user)
	if rerr != nil {
		c.JSON(rerr.StatusCode, gin.H{"error": rerr.Message})
		return
	}

	if network == nil {
		response_utils.RespondWithError(c, http.StatusUnauthorized, "User does not belong to a network")
		return
	}

	c.JSON(http.StatusOK, network)
}

// GetNetworkEvents is a method that returns all network events.
func (mc *ManagerController) GetNetworkEvents(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	events, rerr := mc.service.GetNetworkEvents(&user)
	if rerr != nil {
		c.JSON(rerr.StatusCode, gin.H{"error": rerr.Message})
		return
	}

	c.JSON(http.StatusOK, events)
}
