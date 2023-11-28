package test_service

import (
	"testing"

	"github.com/DowLucas/gin-ticket-release/pkg/database/factory"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/tests/testutils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

var err error
var organizationService *services.OrganisationService

type OrganizationServiceTestSuite struct {
	suite.Suite
	db                  *gorm.DB
	organizationService *services.OrganisationService
}

func createDefaultRole(db *gorm.DB) {
	db.Create(factory.NewRole("validRoleName"))
}

func createDefaultUser(db *gorm.DB) {
	db.Create(factory.NewUser("validUserUGKthID", "validUsername", "validFirstName", "validLastName", "validEmail", 1))
}

func createDefaultOrganization(db *gorm.DB) {
	db.Create(factory.NewOrganization("validOrganizationName"))
}

func (suite *OrganizationServiceTestSuite) SetupTest() {
	println("SetupTest")
	var err error
	suite.db, err = testutils.SetupTestDatabase(false)
	suite.Require().NoError(err)

	createDefaultRole(suite.db)
	createDefaultUser(suite.db)
	createDefaultOrganization(suite.db)

	suite.organizationService = services.NewOrganizationService(suite.db)
}

func (suite *OrganizationServiceTestSuite) TearDownTest() {
	testutils.CleanupTestDatabase(suite.db)
}

func (suite *OrganizationServiceTestSuite) TestAddUserToOrganization() {
	organizationService = services.NewOrganizationService(suite.db)

	// Test adding a valid user to an organization
	err = organizationService.AddUserToOrganization("validUserUGKthID", 1, models.OrgRole("Owner"))
	suite.NoError(err) // Replaces assert.Nil(t, err)

	// Test adding a user that does not exist
	err = organizationService.AddUserToOrganization("invalidUserUGKthID", 1, models.OrgRole("Owner"))
	suite.Error(err) // Replaces assert.NotNil(t, err)

	// Test adding a user to an organization that does not exist
	err = organizationService.AddUserToOrganization("validUserUGKthID", 999, models.OrgRole("Owner"))
	suite.Error(err) // Replaces assert.NotNil(t, err)
}

func (suite *OrganizationServiceTestSuite) TestRemoveUserFromOrganization() {

	organizationService = services.NewOrganizationService(suite.db)

	// Test removing a valid user from an organization
	err = organizationService.RemoveUserFromOrganization("validUserUGKthID", 1)
	suite.NoError(err)

	// Test removing a user that does not exist
	err = organizationService.RemoveUserFromOrganization("invalidUserUGKthID", 1)
	suite.Error(err)

	// Test removing a user from an organization that does not exist
	err = organizationService.RemoveUserFromOrganization("validUserUGKthID", 999)
	suite.Error(err)

}

func (suite *OrganizationServiceTestSuite) TestGetOrganizationUsers() {

	organizationService = services.NewOrganizationService(suite.db)

	// Add a user to an organization
	organizationService.AddUserToOrganization("validUserUGKthID", 1, models.OrgRole("Owner"))

	// Test getting users for a valid organization
	users, _ := organizationService.GetOrganizationUsers(1)
	suite.Equal(1, len(users))

	// Test getting users for an organization that does not exist
	users, _ = organizationService.GetOrganizationUsers(999)

	suite.Equal(0, len(users))
}

func TestOrganizationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceTestSuite))
}
