package admin_controllers

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/DowLucas/gin-ticket-release/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListResources(c *gin.Context, db *gorm.DB, model interface{}, modelName string, idColumn string) {
	queryParams, err := utils.GetQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sortParam := c.DefaultQuery("sort", idColumn)
	sortArray := strings.Split(strings.Trim(sortParam, "[]\""), "\",\"")
	var sort, order string
	if len(sortArray) == 2 {
		sort = sortArray[0]
		order = sortArray[1]
	} else {
		sort = idColumn
		order = "asc"
	}

	if modelName == "users" {
		sort = "id"
	}

	if err := db.Order(sort + " " + order).Offset((queryParams.Page - 1) * queryParams.PerPage).Limit(queryParams.PerPage).Find(model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var count int64
	db.Model(model).Count(&count)

	c.Header("X-Total-Count", fmt.Sprintf("%d", count))
	c.Header("Content-Range", fmt.Sprintf("%s %d-%d/%d", modelName, (queryParams.Page-1)*queryParams.PerPage, (queryParams.Page-1)*queryParams.PerPage+reflect.ValueOf(model).Elem().Len()-1, count))

	c.JSON(http.StatusOK, model)
}
