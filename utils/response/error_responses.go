package response_utils

import "github.com/gin-gonic/gin"

func RespondWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
	c.Abort()
}
