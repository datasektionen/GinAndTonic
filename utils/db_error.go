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

func GetDBError(err error) string {
	if err == gorm.ErrRecordNotFound {
		return "Record not found"
	} else if err == gorm.ErrNotImplemented {
		return "Not implemented"
	} else if err == gorm.ErrMissingWhereClause {
		return "Missing WHERE clause in update"
	} else if err == gorm.ErrUnsupportedDriver {
		return "Unsupported driver"
	} else if err == gorm.ErrPrimaryKeyRequired {
		return "Primary key required"
	} else if err == gorm.ErrInvalidTransaction {
		return "Invalid transaction"
	} else if err == gorm.ErrInvalidData {
		return "Invalid data"
	} else if err == gorm.ErrUnsupportedRelation {
		return "Unsupported relation"
	} else if err == gorm.ErrRegistered {
		return "Already registered"
	} else if err == gorm.ErrInvalidField {
		return "Invalid field"
	} else if err == gorm.ErrEmptySlice {
		return "Empty slice found"
	} else if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return "Duplicate key value violates unique constraint"
	} else {
		return "Unknown database error"
	}
}
