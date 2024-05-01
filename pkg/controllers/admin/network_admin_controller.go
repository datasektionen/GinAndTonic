package admin_controllers

import (
	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NetworkController struct {
	DB *gorm.DB
}

func NewNetworkController(db *gorm.DB) *NetworkController {
	return &NetworkController{DB: db}
}

// ListNetworks handles GET requests to retrieve a list of networks
func (ctrl *NetworkController) ListNetworks(c *gin.Context) {
	var networks []models.Network
	ListResources(c, ctrl.DB, &networks, "networks", "id")
}

func (ctrl *NetworkController) GetNetwork(c *gin.Context) {
	var network models.Network
	if err := ctrl.DB.First(&network, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
		return
	}
	c.JSON(http.StatusOK, network)
}

// CreateNetwork handles POST requests to create a new network
func (ctrl *NetworkController) CreateNetwork(c *gin.Context) {
	var network models.Network
	if err := c.ShouldBindJSON(&network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.DB.Create(&network).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, network)
}

// EditNetwork handles PUT requests to update a network
func (ctrl *NetworkController) EditNetwork(c *gin.Context) {
	id := c.Param("id")
	var network models.Network

	if err := ctrl.DB.First(&network, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
		return
	}

	if err := c.ShouldBindJSON(&network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.DB.Save(&network).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, network)
}

// DeleteNetwork handles DELETE requests to delete a network
func (ctrl *NetworkController) DeleteNetwork(c *gin.Context) {
	id := c.Param("id")
	var network models.Network

	if err := ctrl.DB.First(&network, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
		return
	}

	if err := ctrl.DB.Delete(&network).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Network deleted"})
}
