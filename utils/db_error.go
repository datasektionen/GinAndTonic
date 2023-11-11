package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func HandleDBError(c *gin.Context, err error, action string) {
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Record not found: %v", err)})
		c.Abort()
		return
	} else if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Conflict while %s: %v", action, err)})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error %s: %v", action, err)})
		}
		c.Abort()
	}
}
