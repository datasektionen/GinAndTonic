package authentication

import (
	"log"
	"os"
	"time"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

func GenerateToken(ugkthid string, role string) (string, error) {
	log.Println("Generating token for user", ugkthid, "with role", role)
	expirationTime := time.Now().Add(7 * time.Hour * 24)
	claims := &jwt.StandardClaims{
		Subject:   ugkthid,
		ExpiresAt: expirationTime.Unix(),
		Id:        role,
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
		authHeader := c.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			c.Abort()
			return
		}

		claims, _ := token.Claims.(*jwt.StandardClaims)
		c.Set("ugkthid", claims.Subject)
		println(claims.Id)
		c.Set("role", claims.Id)

		c.Next()
	}
}
