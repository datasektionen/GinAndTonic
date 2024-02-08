package controllers

import (
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

type ExternalAuthController struct {
	DB      *gorm.DB
	service *services.ExternalAuthService
}

// NewExternalAuthController creates a new controller with the given database client
func NewExternalAuthController(db *gorm.DB, service *services.ExternalAuthService) *ExternalAuthController {
	return &ExternalAuthController{DB: db, service: service}
}

/*
Some users dont have a kth.se email address, and therefore cannot use the KTH login system.
This endpoint is used to authenticate these users. These users are not able to create events.
But they can still pay for tickets as long as the ticket release is only for external users.
*/

func generateExternalUGKthID() string {
	/*
		Generates a random string of length 10
	*/
	return "external-" + utils.GenerateRandomString(8)
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

func (eac *ExternalAuthController) SignupExternalUser(c *gin.Context) {
	/*
		Handler that creates a new user with the given information.
	*/
	var externalSignupRequest types.ExternalSignupRequest
	if err := c.ShouldBindJSON(&externalSignupRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if err := externalSignupRequest.Validate(); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	if err := eac.DB.Where("email = ?", externalSignupRequest.Email).First(&models.User{}).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}

	newUGKthID := generateExternalUGKthID()
	// Scramble user firstname and lastname to create a username
	// This is done to avoid username conflicts
	username := generateExternalUsername(externalSignupRequest.FirstName, externalSignupRequest.LastName)

	pwHash, err := utils.HashPassword(externalSignupRequest.Password)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var role models.Role
	if err := eac.DB.Where("name = ?", "external").First(&role).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// generate verify email token
	verifyEmailToken, err := utils.GenerateSecretToken()

	currentTime := time.Now()

	var user models.User = models.User{
		UGKthID:                 newUGKthID,
		Username:                username,
		FirstName:               externalSignupRequest.FirstName,
		LastName:                externalSignupRequest.LastName,
		Email:                   strings.ToLower(externalSignupRequest.Email),
		PasswordHash:            &pwHash,
		IsExternal:              true,
		Role:                    role,
		VerifiedEmail:           false,
		EmailVerificationToken:  verifyEmailToken,
		EmailVerificationSentAt: &currentTime,
	}

	err = models.CreateUserIfNotExist(eac.DB, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	services.Notify_ExternalUserSignupVerification(eac.DB, &user)

	c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

// LoginExternalUser authenticates an external user and returns a token
func (eac *ExternalAuthController) LoginExternalUser(c *gin.Context) {
	/*
		Handler that authenticates an external user and returns a token
	*/
	var loginRequest types.ExternalLoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user
	var user models.User
	if err := eac.DB.Preload("Role").Where("email = ?", strings.ToLower(loginRequest.Email)).
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
	tokenString, err := authentication.GenerateToken(user.UGKthID, user.Role.Name)

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
func (eac *ExternalAuthController) VerifyEmail(c *gin.Context) {
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
