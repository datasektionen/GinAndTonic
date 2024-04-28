package controllers

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CustomerAuthController struct {
	DB      *gorm.DB
	service *services.CustomerAuthService
}

// NewCustomerAuthController creates a new controller with the given database client
func NewCustomerAuthController(db *gorm.DB, service *services.CustomerAuthService) *CustomerAuthController {
	return &CustomerAuthController{DB: db, service: service}
}

/*
Some users don't have a kth.se email address, and therefore cannot use the KTH login system.
This endpoint is used to authenticate these users. These users are not able to create events.
But they can still pay for tickets as long as the ticket release is only for external users.
*/

func generateExternalUGKthID() string {
	/*
		Generates a random string of length 10
	*/
	return "customer-" + utils.GenerateRandomString(8)
}

func scramble(s string) string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	runes := []rune(s)
	for i := len(runes) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func generateExternalUsername(firstName string, lastName string) string {
	firstName = strings.ToLower(firstName)
	lastName = strings.ToLower(lastName)

	// remove spaces
	firstName = strings.ReplaceAll(firstName, " ", "")

	scrambledName := scramble(firstName + lastName)

	return scrambledName
}

func (eac *CustomerAuthController) SignupCustomerUser(c *gin.Context) {
	var err error
	/*
		Handler that creates a new user with the given information.
	*/
	var externalSignupRequest types.CustomerSignupRequest
	if err := c.ShouldBindJSON(&externalSignupRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if err := externalSignupRequest.Validate(); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	var existingUsers []models.User
	if err := eac.DB.Preload("Roles").Where("email = ?", strings.ToLower(externalSignupRequest.Email)).Find(&existingUsers).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	// If any of these users are not a customer_guest, then the email is already in use
	if len(existingUsers) > 0 {
		for _, existingUser := range existingUsers {
			// Basically if the role type is customer it means that the account has not bee saved
			// And cannot be logged in to, so the email can be used again.

			if !existingUser.IsRole(models.RoleCustomer) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
				return
			}
		}
	}

	newUGKthID := generateExternalUGKthID()
	// Scramble user firstname and lastname to create a username
	// This is done to avoid username conflicts

	// The role is either customer or customer_guest
	// Depending on the role, the user can either login or not
	// Same thing goes for the password
	var roleName models.RoleType = models.RoleCustomerGuest
	var pwHash *string = nil
	if externalSignupRequest.IsSaved {
		roleName = models.RoleCustomer

		hash, err := utils.HashPassword(*externalSignupRequest.Password)
		pwHash = &hash
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	var roles []models.Role
	role, err := models.GetRole(eac.DB, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	roles = append(roles, role)

	// generate verify email token
	// This is also only done if the user is saved
	var verifyEmailToken *string = nil
	if externalSignupRequest.IsSaved {
		token, err := utils.GenerateSecretToken()
		verifyEmailToken = &token
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	currentTime := time.Now()

	var user models.User = models.User{
		UGKthID:                 newUGKthID,
		FirstName:               externalSignupRequest.FirstName,
		LastName:                externalSignupRequest.LastName,
		Email:                   strings.ToLower(externalSignupRequest.Email),
		PasswordHash:            pwHash,
		Roles:                   roles,
		VerifiedEmail:           roleName == models.RoleCustomerGuest,
		EmailVerificationToken:  verifyEmailToken,
		EmailVerificationSentAt: &currentTime,
	}

	var requestToken string = ""
	if !externalSignupRequest.IsSaved {
		requestToken, err = utils.GenerateSecretToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		user.RequestToken = &requestToken
	}

	err = models.CreateUserIfNotExist(eac.DB, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if externalSignupRequest.IsSaved {
		services.Notify_ExternalUserSignupVerification(eac.DB, &user)
	} else {
		// Do something else
	}

	c.JSON(http.StatusCreated, gin.H{"user": user, "request_token": requestToken})
}

// LoginExternalUser authenticates an external user and returns a token
func (eac *CustomerAuthController) LoginCustomerUser(c *gin.Context) {
	/*
		Handler that authenticates an external user and returns a token
	*/
	var loginRequest types.CustomerLoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user
	var user models.User
	if err := eac.DB.Joins("JOIN roles ON users.role_id = roles.id").Where("email = ? AND roles.name = ?", strings.ToLower(loginRequest.Email), models.RoleCustomer).
		First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	if !user.VerifiedEmail {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not verified"})
		return
	}

	// Check the password
	if !utils.CheckPasswordHash(loginRequest.Password, *user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Generate a token
	tokenString, err := authentication.GenerateToken(user.UGKthID, user.Roles)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	setCookie(c, tokenString, 60*60*24*7) //  7 days

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

// VerifiedEmailBody is the body of the request to verify an email
type VerifiedEmailBody struct {
	Token string `json:"token"`
}

// VerifyEmail verifies the email of an external user
func (eac *CustomerAuthController) VerifyEmail(c *gin.Context) {
	/*
		Handler that verifies the email of an external user
	*/
	// get token from body
	var body VerifiedEmailBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user via the token
	var user models.User
	if err := eac.DB.Where("email_verification_token = ?", body.Token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	if user.VerifiedEmail {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already verified"})
		return
	}

	// Update the user
	if err := eac.DB.Model(&user).Update("verified_email", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified"})
}

type ResendVerificationEmailBody struct {
	Email string `json:"email"`
}

func (eac *CustomerAuthController) ResendVerificationEmail(c *gin.Context) {
	/*
		Handler that resends the verification email to an external user
	*/
	var body ResendVerificationEmailBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user via the token
	var user models.User
	if err := eac.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	if user.VerifiedEmail {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already verified"})
		return
	}

	services.Notify_ExternalUserSignupVerification(eac.DB, &user)

	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent"})
}
