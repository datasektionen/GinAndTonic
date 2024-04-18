package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrganisationController struct {
	DB                  *gorm.DB
	OrganisationService *services.OrganisationService
}

func NewTeamController(db *gorm.DB, os *services.OrganisationService) *OrganisationController {
	return &OrganisationController{DB: db, OrganisationService: os}
}

func (ec *OrganisationController) CreateTeam(c *gin.Context) {
	var team models.Team

	if err := c.ShouldBindJSON(&team); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "There was an error creating the team"})
		return
	}

	if err := team.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if team already exists
	if ec.DB.Where("name = ?", team.Name).First(&team).RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team already exists"})
		return
	}

	// Check if email is in use
	if ec.DB.Where("email = ?", team.Email).First(&team).RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
		return
	}

	if err := ec.DB.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": utils.GetDBError(err)})
		return
	}

	createdByUserUGKthID, exists := c.Get("ugkthid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get username
	var user models.User
	if err := ec.DB.Where("ug_kth_id = ?", createdByUserUGKthID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": utils.GetDBError(err)})
		return
	}

	ec.OrganisationService.AddUserToTeam(user.Username, team.ID, models.TeamOwner)

	c.JSON(http.StatusCreated, gin.H{"team": team})
}

func (ec *OrganisationController) ListTeams(c *gin.Context) {
	var teams []models.Team

	if err := ec.DB.Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (ec *OrganisationController) ListMyTeams(c *gin.Context) {
	ugkthid, exists := c.Get("ugkthid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := models.GetUserByUGKthIDIfExist(ec.DB, ugkthid.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	teams := user.Teams

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (ec *OrganisationController) GetTeam(c *gin.Context) {
	var team models.Team
	id := c.Param("teamID")

	if err := ec.DB.Preload("Events").
		First(&team, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
}

type UpdateTeamRequest struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (ec *OrganisationController) UpdateTeam(c *gin.Context) {
	var req UpdateTeamRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var team models.Team
	id := c.Param("teamID")

	ID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	if err := ec.DB.First(&team, ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if req.Email != team.Email {
		var existingTeam models.Team
		if ec.DB.Where("email = ?  ", req.Email).First(&existingTeam).RowsAffected > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
			return
		}
	}

	if req.Name != team.Name {
		var existingTeam models.Team
		if ec.DB.Where("name = ?  ", req.Name).First(&existingTeam).RowsAffected > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name already in use"})
			return
		}
	}

	team.Email = req.Email
	team.Name = req.Name

	if err := ec.DB.Save(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
}

func (ec *OrganisationController) DeleteTeam(c *gin.Context) {
	var team models.Team
	id := c.Param("teamID")

	if err := ec.DB.First(&team, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if err := ec.DB.Delete(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
}

func (oc *OrganisationController) ListTeamEvents(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("teamID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	events, err := models.GetAllTeamEvents(oc.DB, uint(teamID))

	c.JSON(http.StatusOK, gin.H{"events": events})
}
