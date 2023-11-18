package services

import (
	"testing"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/stretchr/testify/assert"
)

var organizationService *services.OrganisationService

func SetupOrganizationDB() {
	// Create an organization
	organization := models.Organization{
		Name: "validOrganizationName",
	}
	db.Create(&organization)

	// Create a user
	user := models.User{
		UGKthID:   "validUserUGKthID",
		FirstName: "validUserFirstName",
		LastName:  "validUserLastName",
		Email:     "validUserEmail@gmail.com",
		Username:  "validUserUsername",
		RoleID:    1,
	}
	db.Create(&user)
	organizationService = services.NewOrganizationService(db)
	return
}

func TestAddUserToOrganization(t *testing.T) {
	// Test adding a valid user to an organization
	err := organizationService.AddUserToOrganization("validUserUGKthID", 1, models.OrgRole("Owner"))
	assert.Nil(t, err)

	// Test adding a user that does not exist
	err = organizationService.AddUserToOrganization("invalidUserUGKthID", 1, models.OrgRole("Owner"))
	assert.NotNil(t, err)

	// Test adding a user to an organization that does not exist
	err = organizationService.AddUserToOrganization("validUserUGKthID", 999, models.OrgRole("Owner"))
	assert.NotNil(t, err)
}

func TestRemoveUserFromOrganization(t *testing.T) {

	// Test removing a valid user from an organization
	err := organizationService.RemoveUserFromOrganization("validUserUGKthID", 1)
	assert.Nil(t, err)

	// Test removing a user that does not exist
	err = organizationService.RemoveUserFromOrganization("invalidUserUGKthID", 1)
	assert.NotNil(t, err)

	// Test removing a user from an organization that does not exist
	err = organizationService.RemoveUserFromOrganization("validUserUGKthID", 999)
	assert.NotNil(t, err)

}

func TestGetOrganizationUsers(t *testing.T) {

	// Test getting users for a valid organization
	users, _ := organizationService.GetOrganizationUsers(1)
	assert.NotEmpty(t, users)

	// Test getting users for an organization that does not exist
	users, _ = organizationService.GetOrganizationUsers(999)

	assert.Empty(t, users)

	// Additional test cases...
}
