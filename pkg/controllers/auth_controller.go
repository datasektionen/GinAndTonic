package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/DowLucas/gin-ticket-release/pkg/authentication"
	"github.com/DowLucas/gin-ticket-release/pkg/database"
	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

const YOUR_CALLBACK_URL = "http://localhost:8080/login-complete"

var db *gorm.DB
var jwtKey []byte

func init() {
	var err error

	if err = godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
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
	c.Redirect(http.StatusSeeOther, "/")
}

// Login redirects to the external login page
func Login(c *gin.Context) {
	println(os.Getenv("LOGIN_URL"))
	loginURL := os.Getenv("LOGIN_BASE_URL") + "/login?callback=" + YOUR_CALLBACK_URL
	println(loginURL)
	c.Redirect(http.StatusSeeOther, loginURL)
}

// LoginComplete handles the callback from the external login system
func LoginComplete(c *gin.Context) {
	token := c.Param("token")
	client := &http.Client{}

	url := os.Getenv("LOGIN_BASE_URL") + "/verify/" + token + ".json?api_key=" + os.Getenv("LOGIN_API_KEY")
	println(url)

	req, err := http.NewRequest("GET", url, nil)
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
			tokenString, err := authentication.GenerateToken(body.UGKthID, user.Role.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Login successful!",
				"user":    user,
				"token":   tokenString,
			})
			return
		}

		role, err := models.GetRole(db, "user")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error fetching role",
			})
			return
		}

		user = models.User{
			Username:  body.User,
			FirstName: body.FirstName,
			LastName:  body.LastName,
			Email:     body.Emails,
			UGKthID:   body.UGKthID,
			RoleID:    role.ID,
			Role:      role,
		}

		tokenString, err := authentication.GenerateToken(body.UGKthID, user.Role.Name)
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

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful!",
			"user":    user,
			"token":   tokenString,
		})

	} else {
		println("Error: " + res.Status)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error logging in",
		})
	}
}
