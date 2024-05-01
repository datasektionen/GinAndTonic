package authentication

import (
	"fmt"
	"os"
	"time"

	"net/http"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateToken(user_id string, roles []models.Role) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = string(role.Name)
	}

	claims := &Claims{
		UserID: user_id,
		Roles:  roleNames,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateTokenMiddleware(
	failOnError bool,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Print cookie
		cookie, err := c.Request.Cookie("auth_token")
		// View cookie error
		if err != nil {

			if !failOnError {
				c.Next()
				return
			}
			fmt.Println("Error getting cookie:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Print cookie
		tokenString := cookie.Value

		if tokenString == "" {
			if !failOnError {
				c.Next()
				return
			}

			fmt.Println("Error getting cookie:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			c.Abort()
			return
		}

		claims, _ := token.Claims.(*Claims)

		c.Set("user_id", claims.UserID)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}
