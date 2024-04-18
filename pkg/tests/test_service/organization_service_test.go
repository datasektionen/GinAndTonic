package test_service

import (
	"fmt"
	"testing"

	"github.com/DowLucas/gin-ticket-release/pkg/database/factory"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/tests/testutils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

var err error
var teamService *services.OrganisationService

type TeamServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	teamService *services.OrganisationService
}

func createDefaultRole(db *gorm.DB) {
	var role *models.Role
	role = factory.NewRole("validRoleName")

	err := db.Create(*role)
	if err != nil {
		fmt.Println("Error creating role", err)
		panic(err)
	}
}

func createDefaultUser(db *gorm.DB) {
	err := db.Create(factory.NewUser("validUserUGKthID", "validUsername", "validFirstName", "validLastName", "validEmail", 1))
	if err != nil {
		fmt.Println("Error creating role", err)
		panic(err)
	}
}

func createDefaultTeam(db *gorm.DB) {
	err := db.Create(factory.NewTeam("validTeamName", "validTeamEmail"))
	if err != nil {
		panic(err)
	}

}

func (suite *TeamServiceTestSuite) SetupTest() {
	println("SetupTest")
	var err error
	suite.db, err = testutils.SetupTestDatabase(false)
	suite.Require().NoError(err)

	createDefaultRole(suite.db)
	createDefaultUser(suite.db)
	createDefaultTeam(suite.db)

	suite.teamService = services.NewTeamService(suite.db)
}

func (suite *TeamServiceTestSuite) TearDownTest() {
	testutils.CleanupTestDatabase(suite.db)
}

func (suite *TeamServiceTestSuite) TestAddUserToTeam() {
	teamService = services.NewTeamService(suite.db)

	// Test adding a valid user to an team
	err = teamService.AddUserToTeam("validUserUGKthID", 1, models.OrgRole("Owner"))
	suite.NoError(err) // Replaces assert.Nil(t, err)

	// Test adding a user that does not exist
	err = teamService.AddUserToTeam("invalidUserUGKthID", 1, models.OrgRole("Owner"))
	suite.Error(err) // Replaces assert.NotNil(t, err)

	// Test adding a user to an team that does not exist
	err = teamService.AddUserToTeam("validUserUGKthID", 999, models.OrgRole("Owner"))
	suite.Error(err) // Replaces assert.NotNil(t, err)
}

func (suite *TeamServiceTestSuite) TestRemoveUserFromTeam() {

	teamService = services.NewTeamService(suite.db)

	// Test removing a valid user from an team
	err = teamService.RemoveUserFromTeam("validUserUGKthID", 1)
	suite.NoError(err)

	// Test removing a user that does not exist
	err = teamService.RemoveUserFromTeam("invalidUserUGKthID", 1)
	suite.Error(err)

	// Test removing a user from an team that does not exist
	err = teamService.RemoveUserFromTeam("validUserUGKthID", 999)
	suite.Error(err)

}

func (suite *TeamServiceTestSuite) TestGetTeamUsers() {

	teamService = services.NewTeamService(suite.db)

	// Add a user to an team
	teamService.AddUserToTeam("validUserUGKthID", 1, models.OrgRole("Owner"))

	// Test getting users for a valid team
	users, _ := teamService.GetTeamUsers(1)
	suite.Equal(1, len(users))

	// Test getting users for an team that does not exist
	users, _ = teamService.GetTeamUsers(999)

	suite.Equal(0, len(users))
}

func TestTeamServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeamServiceTestSuite))
}
