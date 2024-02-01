package authentication

import (
	"fmt"
	"os"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

func GenerateToken(ugkthid string, role string) (string, error) {
	expirationTime := time.Now().Add(7 * time.Hour * 24)
	claims := &jwt.RegisteredClaims{
		Subject:   ugkthid,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		ID:        role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Print cookie
		cookie, err := c.Request.Cookie("auth_token")
		// View cookie error
		if err != nil {
			fmt.Println("Error getting cookie:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Print cookie
		tokenString := cookie.Value

		if tokenString == "" {
			fmt.Println("Error getting cookie:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			c.Abort()
			return
		}

		claims, _ := token.Claims.(*jwt.RegisteredClaims)
		c.Set("ugkthid", claims.Subject)
		c.Set("role", claims.ID)

		c.Next()
	}
}
