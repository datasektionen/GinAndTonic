package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/database"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var db *gorm.DB
var jwtKey []byte

func getSameSite() http.SameSite {
	if os.Getenv("ENV") == "dev" {
		return http.SameSiteNoneMode
	} else if os.Getenv("ENV") == "prod" {
		return http.SameSiteNoneMode
	}
	return http.SameSiteNoneMode
}

func getDomain() string {
	if os.Getenv("ENV") == "dev" {
		return "localhost"
	} else if os.Getenv("ENV") == "prod" {
		return ".datasektionen.se"
	}

	return "localhost"
}

func setCookie(c *gin.Context, tokenString string, maxAge int) {
	if os.Getenv("ENV") == "dev" {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			HttpOnly: true, // Set this to true in production
			Path:     "/",
		})
	} else if os.Getenv("ENV") == "prod" {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			HttpOnly: true, // Set this to true in production
			Path:     "/",
			MaxAge:   maxAge,
			SameSite: getSameSite(),
			Secure:   os.Getenv("ENV") == "prod", // True means only send cookie over HTTPS
			Domain:   getDomain(),                // Set your domain here
		})
	}
}

func init() {
	var err error

	if os.Getenv("ENV") == "dev" {
		if err = godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	jwtKey = []byte(os.Getenv("JWT_KEY"))

	db, err = database.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	db.AutoMigrate(&models.User{})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	// Logout logic
	// Remove the cookie
	setCookie(c, "", -1)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// Login redirects to the external login page
func LoginPostman(c *gin.Context) {
	c.Redirect(http.StatusSeeOther, "http://localhost:7002/login?callback="+"http://localhost:8080/postman-login-complete/")
}

func Login(c *gin.Context) {
	// loginURL := os.Getenv("LOGIN_BASE_URL") + "/login?callback=" + "http://app:8080/login-complete/"
	if os.Getenv("ENV") == "dev" {
		c.JSON(http.StatusOK, gin.H{
			"login_url": "http://localhost:7002/login?callback=" + "http://localhost:8080/login-complete/",
		})

	} else if os.Getenv("ENV") == "prod" {
		// Redirect to the external login page
		scheme := "https" // Set this to "http" if your application is not running on HTTPS
		callbackURL := scheme + "://" + c.Request.Host + "/login-complete/"
		c.JSON(http.StatusOK, gin.H{
			"login_url": os.Getenv("LOGIN_BASE_URL") + "/login?callback=" + callbackURL,
		})
	}
}

func CurrentUser(c *gin.Context) {
	// Get the user from the context
	user_id := c.MustGet("user_id").(string)

	fmt.Println("User ID: ", user_id)

	// Get the user from the database
	user, err := models.GetUserByUGKthIDIfExist(db, user_id)

	if err != nil {
		// Remove the cookie
		setCookie(c, "", -1)

		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// LoginComplete handles the callback from the external login system
func LoginComplete(c *gin.Context) {
	token := c.Param("token")
	client := &http.Client{}

	verificationURL := os.Getenv("LOGIN_BASE_URL") + "/verify/" + token + ".json?api_key=" + os.Getenv("LOGIN_API_KEY")

	req, err := http.NewRequest("GET", verificationURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating request",
		})
		return
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api_key", os.Getenv("LOGIN_API_KEY"))
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error sending request",
		})
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var body types.Body
		decoder := json.NewDecoder(res.Body)
		err := decoder.Decode(&body)
		if err != nil {
			println("Error: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error decoding response",
			})
			return
		}

		// Check if user exists in database
		var user models.User
		user, err = models.GetUserByUGKthIDIfExist(db, body.UGKthID)
		if err == nil {
			// User exists
			tokenString, err := authentication.GenerateToken(body.UGKthID, user.Roles)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			setCookie(c, tokenString, 60*60*24*7) //  7 days

			// referer := c.Request.Referer()
			c.Redirect(http.StatusSeeOther, os.Getenv("FRONTEND_BASE_URL")+"/handle-login-callback?auth=success")

			return
		}

		// User does not exist

		var roles []models.Role
		if body.Emails == "turetek@kth.se" || body.Emails == "dow@kth.se" { // Given super_admin role
			roles = append(roles, models.Role{Name: models.RoleSuperAdmin})
		} else {
			roles = append(roles, models.Role{Name: models.RoleCustomer})
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error fetching role",
			})
			return
		}

		user = models.User{
			FirstName:     body.FirstName,
			LastName:      body.LastName,
			Email:         body.Emails,
			UGKthID:       body.UGKthID,
			Roles:         roles,
			VerifiedEmail: true,
		}

		tokenString, err := authentication.GenerateToken(body.UGKthID, user.Roles)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = models.CreateUserIfNotExist(db, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user food preference",
			})
			return
		}

		// Set the JWT token in an HTTP-only cookie

		setCookie(c, tokenString, 60*60*24*7) //  7 days

		services.Notify_Welcome(db, &user)

		c.Redirect(http.StatusSeeOther, os.Getenv("FRONTEND_BASE_URL")+"/handle-login-callback?auth=success")
	} else {
		println("Error: " + res.Status)
		http.Redirect(c.Writer, c.Request, os.Getenv("FRONTEND_BASE_URL")+"/handle-login-callback?auth=failed", http.StatusSeeOther)
	}
}

// LoginCompletePostman handles the callback from the external login system
func LoginCompletePostman(c *gin.Context) {
	token := c.Param("token")
	client := &http.Client{}

	verificationURL := os.Getenv("LOGIN_BASE_URL") + "/verify/" + token + ".json?api_key=" + os.Getenv("LOGIN_API_KEY")

	req, err := http.NewRequest("GET", verificationURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating request",
		})
		return
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api_key", os.Getenv("LOGIN_API_KEY"))
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error sending request",
		})
		return
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var body types.Body
		decoder := json.NewDecoder(res.Body)
		err := decoder.Decode(&body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error decoding response", "error": err.Error(),
			})
			return
		}

		// Check if user exists in database
		var user models.User
		user, err = models.GetUserByUGKthIDIfExist(db, body.UGKthID)
		if err == nil {
			// User exists
			tokenString, err := authentication.GenerateToken(body.UGKthID, user.Roles)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			setCookie(c, tokenString, 60*60*24*7) //  7 days

			c.JSON(http.StatusOK, gin.H{
				"token": tokenString,
				"user":  user,
			})

			return
		}

		var roles []models.Role
		if body.Emails == "turetek@kth.se" || body.Emails == "dow@kth.se" {
			// Given super_admin role
			roles = append(roles, models.Role{Name: models.RoleSuperAdmin})
		} else {
			roles = append(roles, models.Role{Name: models.RoleCustomer})
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error fetching role",
			})
			return
		}

		user = models.User{
			FirstName:     body.FirstName,
			LastName:      body.LastName,
			Email:         body.Emails,
			UGKthID:       body.UGKthID,
			Roles:         roles,
			VerifiedEmail: true,
		}

		tokenString, err := authentication.GenerateToken(body.UGKthID, user.Roles)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = models.CreateUserIfNotExist(db, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user food preference",
			})
			return
		}

		setCookie(c, tokenString, 60*60*24*7) //  7 days

		// Set the JWT token in an HTTP-only cookie
		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"user":  user,
		})

	} else {
		println("Error: " + res.Status)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error verifying token",
		})
	}
}
